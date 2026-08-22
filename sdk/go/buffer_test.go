package datasnoop_test

import (
	"context"
	"errors"
	"github.com/datasnoop/datasnoop/sdk/go"
	"sync"
	"testing"
	"time"
)

func TestBufferNeverBlocksAndCountsDiscards(t *testing.T) {
	block := make(chan struct{})
	started := make(chan struct{})
	var startOnce sync.Once
	buffer, err := datasnoop.NewBuffer(datasnoop.BufferConfig{Capacity: 1, Attempts: 1, Timeout: time.Second}, func(context.Context, any) error {
		startOnce.Do(func() { close(started) })
		<-block
		return errors.New("unavailable")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !buffer.Submit("first") {
		t.Fatal("first submit should fit")
	}
	<-started
	if !buffer.Submit("second") {
		t.Fatal("second submit should fit queue")
	}
	if buffer.Submit("third") {
		t.Fatal("saturated submit must discard")
	}
	close(block)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := buffer.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if buffer.Stats().Dropped < 3 {
		t.Fatalf("stats = %#v", buffer.Stats())
	}
}
func TestBufferShutdownHonorsDeadline(t *testing.T) {
	block := make(chan struct{})
	buffer, err := datasnoop.NewBuffer(datasnoop.BufferConfig{Capacity: 1, Attempts: 1, Timeout: time.Second}, func(context.Context, any) error { <-block; return nil })
	if err != nil {
		t.Fatal(err)
	}
	buffer.Submit("record")
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if !errors.Is(buffer.Shutdown(ctx), context.DeadlineExceeded) {
		t.Fatal("expected shutdown deadline")
	}
	close(block)
}
