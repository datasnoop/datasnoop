package retention

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrWorkBudgetExhausted = errors.New("retention database work budget is exhausted")

// Executor is the narrow database boundary used by retention work.
type Executor interface {
	Exec(context.Context, string, ...any) error
}

type poolExecutor struct{ pool *pgxpool.Pool }

func (executor poolExecutor) Exec(ctx context.Context, query string, arguments ...any) error {
	_, err := executor.pool.Exec(ctx, query, arguments...)
	return err
}

// Cycle records the result of one bounded retention run.
type Cycle struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Failure    error
}

// Runner expires full TimescaleDB chunks older than the active policy and records every outcome.
type Runner struct {
	database Executor
	policy   func() Policy
	now      func() time.Time
	budget   *Budget
}

func NewRunner(database Executor, policy func() Policy) *Runner {
	return &Runner{database: database, policy: policy, now: time.Now}
}

// NewPoolRunner connects retention work to the TimescaleDB pool without exposing it to callers.
func NewPoolRunner(pool *pgxpool.Pool, policy func() Policy) *Runner {
	return NewRunner(poolExecutor{pool: pool}, policy)
}

// WithBudget limits concurrent retention database work independently from ingestion.
func (runner *Runner) WithBudget(budget *Budget) *Runner {
	runner.budget = budget
	return runner
}

func (runner *Runner) Run(ctx context.Context) Cycle {
	started := runner.now().UTC()
	cycle := Cycle{StartedAt: started}
	if runner.budget != nil && !runner.budget.TryAcquire() {
		cycle.Failure = ErrWorkBudgetExhausted
		cycle.FinishedAt = runner.now().UTC()
		cycle.Duration = cycle.FinishedAt.Sub(cycle.StartedAt)
		runner.record(ctx, &cycle)
		return cycle
	}
	if runner.budget != nil {
		defer runner.budget.Release()
	}
	cutoff := started.Add(-runner.policy().Duration)
	for _, relation := range []string{"operations", "logs", "metrics"} {
		if err := runner.database.Exec(ctx, "SELECT drop_chunks(relation => $1::regclass, older_than => $2::timestamptz)", relation, cutoff); err != nil {
			cycle.Failure = fmt.Errorf("expire %s: %w", relation, err)
			break
		}
	}
	cycle.FinishedAt = runner.now().UTC()
	cycle.Duration = cycle.FinishedAt.Sub(cycle.StartedAt)
	runner.record(ctx, &cycle)
	return cycle
}

func (runner *Runner) record(ctx context.Context, cycle *Cycle) {
	if cycle.Failure != nil {
		_ = runner.database.Exec(ctx, "INSERT INTO retention_cycles (started_at, finished_at, duration_ms, retention_interval, status, failure) VALUES ($1,$2,$3,$4,'failed',$5)", cycle.StartedAt, cycle.FinishedAt, cycle.Duration.Milliseconds(), runner.policy().Duration.String(), cycle.Failure.Error())
		return
	}
	if err := runner.database.Exec(ctx, "INSERT INTO retention_cycles (started_at, finished_at, duration_ms, retention_interval, status) VALUES ($1,$2,$3,$4,'completed')", cycle.StartedAt, cycle.FinishedAt, cycle.Duration.Milliseconds(), runner.policy().Duration.String()); err != nil {
		cycle.Failure = fmt.Errorf("record retention cycle: %w", err)
	}
}

// Budget is an independent bounded budget for destructive retention database work.
type Budget struct{ slots chan struct{} }

func NewBudget(limit int) (*Budget, error) {
	if limit < 1 {
		return nil, errors.New("retention work budget must be positive")
	}
	return &Budget{slots: make(chan struct{}, limit)}, nil
}

func (budget *Budget) TryAcquire() bool {
	select {
	case budget.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (budget *Budget) Release() { <-budget.slots }

// Schedule runs cycles at the supplied interval until the returned stop function is called.
func (runner *Runner) Schedule(ctx context.Context, interval time.Duration) (func(), error) {
	if interval < time.Minute {
		return nil, fmt.Errorf("retention schedule interval must be at least one minute")
	}
	child, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-child.Done():
				return
			case <-ticker.C:
				runner.Run(child)
			}
		}
	}()
	return cancel, nil
}
