package observe

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sidwiskers/hermes/api"
)

// Metrics is a fixed-cardinality, lock-free aggregate observer. It implements
// both UpdateObserver and api.Observer.
type Metrics struct {
	updatesStarted   atomic.Uint64
	updatesSucceeded atomic.Uint64
	updatesFailed    atomic.Uint64
	updatesPanicked  atomic.Uint64
	updatesInFlight  atomic.Int64
	updatesPeak      atomic.Int64
	updateNanos      atomic.Uint64

	callsStarted   atomic.Uint64
	callsSucceeded atomic.Uint64
	callsFailed    atomic.Uint64
	callsInFlight  atomic.Int64
	callsPeak      atomic.Int64
	callNanos      atomic.Uint64
}

// Snapshot is a consistent-enough lock-free metrics read. Individual counters
// may advance while the snapshot is collected.
type Snapshot struct {
	UpdatesStarted      uint64
	UpdatesSucceeded    uint64
	UpdatesFailed       uint64
	UpdatesPanicked     uint64
	UpdatesInFlight     int64
	UpdatesPeakInFlight int64
	UpdateTotalDuration time.Duration
	CallsStarted        uint64
	CallsSucceeded      uint64
	CallsFailed         uint64
	CallsInFlight       int64
	CallsPeakInFlight   int64
	CallTotalDuration   time.Duration
}

// AverageUpdateDuration returns the mean duration of completed updates.
func (s Snapshot) AverageUpdateDuration() time.Duration {
	completed := s.UpdatesSucceeded + s.UpdatesFailed + s.UpdatesPanicked
	if completed == 0 {
		return 0
	}
	return s.UpdateTotalDuration / time.Duration(completed)
}

// AverageCallDuration returns the mean duration of completed Bot API calls.
func (s Snapshot) AverageCallDuration() time.Duration {
	completed := s.CallsSucceeded + s.CallsFailed
	if completed == 0 {
		return 0
	}
	return s.CallTotalDuration / time.Duration(completed)
}

// StartUpdate implements UpdateObserver.
func (m *Metrics) StartUpdate(ctx context.Context, _ UpdateEvent) context.Context {
	if m == nil {
		return ctx
	}
	m.updatesStarted.Add(1)
	inFlight := m.updatesInFlight.Add(1)
	updatePeak(&m.updatesPeak, inFlight)
	return ctx
}

// FinishUpdate implements UpdateObserver.
func (m *Metrics) FinishUpdate(_ context.Context, _ UpdateEvent, result UpdateResult) {
	if m == nil {
		return
	}
	m.updatesInFlight.Add(-1)
	m.updateNanos.Add(uint64(max(result.Duration, 0)))
	switch {
	case result.Panicked:
		m.updatesPanicked.Add(1)
	case result.Err != nil:
		m.updatesFailed.Add(1)
	default:
		m.updatesSucceeded.Add(1)
	}
}

// StartCall implements api.Observer.
func (m *Metrics) StartCall(ctx context.Context, _ api.CallEvent) context.Context {
	if m == nil {
		return ctx
	}
	m.callsStarted.Add(1)
	inFlight := m.callsInFlight.Add(1)
	updatePeak(&m.callsPeak, inFlight)
	return ctx
}

// FinishCall implements api.Observer.
func (m *Metrics) FinishCall(_ context.Context, _ api.CallEvent, result api.CallResult) {
	if m == nil {
		return
	}
	m.callsInFlight.Add(-1)
	m.callNanos.Add(uint64(max(result.Duration, 0)))
	if result.Err != nil {
		m.callsFailed.Add(1)
	} else {
		m.callsSucceeded.Add(1)
	}
}

// Snapshot returns the current aggregate counters.
func (m *Metrics) Snapshot() Snapshot {
	if m == nil {
		return Snapshot{}
	}
	return Snapshot{
		UpdatesStarted:      m.updatesStarted.Load(),
		UpdatesSucceeded:    m.updatesSucceeded.Load(),
		UpdatesFailed:       m.updatesFailed.Load(),
		UpdatesPanicked:     m.updatesPanicked.Load(),
		UpdatesInFlight:     m.updatesInFlight.Load(),
		UpdatesPeakInFlight: m.updatesPeak.Load(),
		UpdateTotalDuration: time.Duration(m.updateNanos.Load()),
		CallsStarted:        m.callsStarted.Load(),
		CallsSucceeded:      m.callsSucceeded.Load(),
		CallsFailed:         m.callsFailed.Load(),
		CallsInFlight:       m.callsInFlight.Load(),
		CallsPeakInFlight:   m.callsPeak.Load(),
		CallTotalDuration:   time.Duration(m.callNanos.Load()),
	}
}

func updatePeak(peak *atomic.Int64, value int64) {
	for current := peak.Load(); value > current; current = peak.Load() {
		if peak.CompareAndSwap(current, value) {
			return
		}
	}
}
