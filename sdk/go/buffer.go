package datasnoop

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Delivery func(context.Context, any) error

type BufferConfig struct {
	Capacity, Attempts int
	Timeout            time.Duration
}
type BufferStats struct{ Dropped, Retried uint64 }

type Buffer struct {
	deliver  Delivery
	queue    chan any
	attempts int
	timeout  time.Duration
	dropped  atomic.Uint64
	retried  atomic.Uint64
	done     chan struct{}
	stop     chan struct{}
	wait     sync.WaitGroup
	stopOnce sync.Once
	doneOnce sync.Once
}

func NewBuffer(config BufferConfig, delivery Delivery) (*Buffer, error) {
	if delivery == nil || config.Capacity < 1 || config.Attempts < 1 || config.Timeout <= 0 {
		return nil, errors.New("buffer requires a delivery function and positive capacity, attempts, and timeout")
	}
	buffer := &Buffer{deliver: delivery, queue: make(chan any, config.Capacity), attempts: config.Attempts, timeout: config.Timeout, done: make(chan struct{}), stop: make(chan struct{})}
	buffer.wait.Go(buffer.run)
	return buffer, nil
}

func (buffer *Buffer) Submit(record any) bool {
	select {
	case <-buffer.stop:
		buffer.dropped.Add(1)
		return false
	default:
	}
	select {
	case buffer.queue <- record:
		return true
	default:
		buffer.dropped.Add(1)
		return false
	}
}
func (buffer *Buffer) Stats() BufferStats {
	return BufferStats{Dropped: buffer.dropped.Load(), Retried: buffer.retried.Load()}
}
func (buffer *Buffer) Shutdown(ctx context.Context) error {
	buffer.stopOnce.Do(func() { close(buffer.stop) })
	buffer.doneOnce.Do(func() { go func() { buffer.wait.Wait(); close(buffer.done) }() })
	select {
	case <-buffer.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (buffer *Buffer) run() {
	for {
		select {
		case record := <-buffer.queue:
			buffer.send(record)
		case <-buffer.stop:
			for {
				select {
				case record := <-buffer.queue:
					buffer.send(record)
				default:
					return
				}
			}
		}
	}
}
func (buffer *Buffer) send(record any) {
	for attempt := 0; attempt < buffer.attempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), buffer.timeout)
		err := buffer.deliver(ctx, record)
		cancel()
		if err == nil {
			return
		}
		if attempt+1 < buffer.attempts {
			buffer.retried.Add(1)
		}
	}
	buffer.dropped.Add(1)
}
