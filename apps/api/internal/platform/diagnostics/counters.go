// Package diagnostics exposes restart-scoped operational telemetry counters.
package diagnostics

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
)

type Snapshot struct {
	RestartID string
	Accepted  uint64
	Rejected  uint64
	Throttled uint64
	Dropped   uint64
}

type Counters struct {
	restartID string
	accepted  atomic.Uint64
	rejected  atomic.Uint64
	throttled atomic.Uint64
	dropped   atomic.Uint64
}

func NewCounters() *Counters { return &Counters{restartID: newRestartID()} }

func (counters *Counters) Accept(count uint64)   { counters.accepted.Add(count) }
func (counters *Counters) Reject(count uint64)   { counters.rejected.Add(count) }
func (counters *Counters) Throttle(count uint64) { counters.throttled.Add(count) }
func (counters *Counters) Drop(count uint64)     { counters.dropped.Add(count) }

func (counters *Counters) Snapshot() Snapshot {
	return Snapshot{RestartID: counters.restartID, Accepted: counters.accepted.Load(), Rejected: counters.rejected.Load(), Throttled: counters.throttled.Load(), Dropped: counters.dropped.Load()}
}

func newRestartID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("cannot generate diagnostics restart identity")
	}
	return hex.EncodeToString(bytes)
}
