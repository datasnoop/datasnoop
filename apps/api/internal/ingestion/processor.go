package ingestion

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrAdmissionFull = errors.New("ingestion admission capacity is exhausted")
var ErrQueueFull = errors.New("ingestion queue capacity is exhausted")
var ErrPersistenceTimeout = errors.New("ingestion persistence deadline exceeded")

type ProcessorConfig struct {
	AdmissionLimit     int
	QueueLimit         int
	Workers            int
	PersistenceTimeout time.Duration
}

type Processor struct {
	store              BatchStore
	admission          chan struct{}
	queue              chan submission
	persistenceTimeout time.Duration
	stopped            chan struct{}
	waitGroup          sync.WaitGroup
}

type submission struct {
	batch  Batch
	result chan Outcome
}

func NewProcessor(config ProcessorConfig, store BatchStore) (*Processor, error) {
	if store == nil || config.AdmissionLimit < 1 || config.QueueLimit < 1 || config.Workers < 1 || config.PersistenceTimeout <= 0 {
		return nil, errors.New("ingestion processor requires positive independent bounds and a batch store")
	}
	processor := &Processor{store: store, admission: make(chan struct{}, config.AdmissionLimit), queue: make(chan submission, config.QueueLimit), persistenceTimeout: config.PersistenceTimeout, stopped: make(chan struct{})}
	for range config.Workers {
		processor.waitGroup.Go(processor.run)
	}
	return processor, nil
}

func (processor *Processor) Persist(ctx context.Context, batch Batch) (Outcome, error) {
	select {
	case processor.admission <- struct{}{}:
		defer func() { <-processor.admission }()
	default:
		return Outcome{}, ErrAdmissionFull
	}
	result := make(chan Outcome, 1)
	select {
	case processor.queue <- submission{batch: batch, result: result}:
	case <-processor.stopped:
		return Outcome{}, errors.New("ingestion processor is stopped")
	default:
		return Outcome{}, ErrQueueFull
	}
	select {
	case outcome := <-result:
		if outcome.Failed != nil && errors.Is(outcome.Failed, context.DeadlineExceeded) {
			return outcome, ErrPersistenceTimeout
		}
		return outcome, outcome.Failed
	case <-ctx.Done():
		return Outcome{}, ctx.Err()
	case <-processor.stopped:
		return Outcome{}, errors.New("ingestion processor is stopped")
	}
}

func (processor *Processor) Close() {
	select {
	case <-processor.stopped:
		return
	default:
		close(processor.stopped)
		processor.waitGroup.Wait()
	}
}

// QueueDepth reports bounded ingest work only; query, retention, and live
// delivery use separate work budgets.
func (processor *Processor) QueueDepth() int { return len(processor.queue) }

func (processor *Processor) run() {
	for {
		select {
		case <-processor.stopped:
			return
		case submission := <-processor.queue:
			writeContext, cancel := context.WithTimeout(context.Background(), processor.persistenceTimeout)
			outcome := processor.store.Persist(writeContext, submission.batch)
			if errors.Is(writeContext.Err(), context.DeadlineExceeded) && outcome.Failed == nil {
				outcome.Failed = context.DeadlineExceeded
			}
			cancel()
			select {
			case submission.result <- outcome:
			case <-processor.stopped:
				return
			}
		}
	}
}
