package metrics

import (
	"sync/atomic"
)

// Metrics holds the metrics for the application
type Metrics struct {
	generated atomic.Int64
	consumed  atomic.Int64
}

// New creates a new Metrics instance
func New() *Metrics {
	return &Metrics{}
}

// IncrementGenerated increments the generated keys counter
func (m *Metrics) IncrementGenerated() {
	m.generated.Add(1)
}

// IncrementConsumed increments the consumed keys counter
func (m *Metrics) IncrementConsumed() {
	m.consumed.Add(1)
}

// GetCounts returns the current counts
func (m *Metrics) GetCounts() (int64, int64) {
	return m.generated.Load(), m.consumed.Load()
}

// Reset resets all counters to zero
func (m *Metrics) Reset() {
	m.generated.Store(0)
	m.consumed.Store(0)
}
