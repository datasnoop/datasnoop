package ingestion_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
)

func TestProcessorReturnsRetryableCapacityErrorsAtIndependentBounds(t *testing.T) {
	t.Parallel()
	store := &blockingStore{started: make(chan struct{}), release: make(chan struct{})}
	processor, err := ingestion.NewProcessor(ingestion.ProcessorConfig{AdmissionLimit: 2, QueueLimit: 1, Workers: 1, PersistenceTimeout: time.Second}, store)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(processor.Close)

	first := make(chan error, 1)
	go func() { _, err := processor.Persist(context.Background(), ingestion.Batch{}); first <- err }()
	<-store.started
	second := make(chan error, 1)
	go func() { _, err := processor.Persist(context.Background(), ingestion.Batch{}); second <- err }()
	deadline := time.After(time.Second)
	for processor.QueueDepth() != 1 {
		select {
		case <-deadline:
			t.Fatal("second submission did not queue")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if _, err := processor.Persist(context.Background(), ingestion.Batch{}); !errors.Is(err, ingestion.ErrAdmissionFull) && !errors.Is(err, ingestion.ErrQueueFull) {
		t.Fatalf("saturated processor error = %v", err)
	}
	close(store.release)
	if err := <-first; err != nil {
		t.Fatalf("first result: %v", err)
	}
	if err := <-second; err != nil {
		t.Fatalf("second result: %v", err)
	}
}

func TestProcessorAppliesPersistenceTimeout(t *testing.T) {
	t.Parallel()
	store := &blockingStore{started: make(chan struct{}), release: make(chan struct{})}
	processor, err := ingestion.NewProcessor(ingestion.ProcessorConfig{AdmissionLimit: 1, QueueLimit: 1, Workers: 1, PersistenceTimeout: 10 * time.Millisecond}, store)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { close(store.release); processor.Close() })
	_, err = processor.Persist(context.Background(), ingestion.Batch{})
	if !errors.Is(err, ingestion.ErrPersistenceTimeout) {
		t.Fatalf("persistence timeout = %v", err)
	}
}

type blockingStore struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (store *blockingStore) Persist(ctx context.Context, _ ingestion.Batch) ingestion.Outcome {
	store.once.Do(func() { close(store.started) })
	select {
	case <-store.release:
		return ingestion.Outcome{}
	case <-ctx.Done():
		return ingestion.Outcome{Failed: ctx.Err()}
	}
}
