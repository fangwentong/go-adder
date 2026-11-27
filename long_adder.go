package adder

import (
	"math/rand/v2"
	"sync/atomic"
)

// LongAdder is a high-concurrency int64 counter inspired by Java's LongAdder.
// Under low contention it updates a single base value; under high contention it lazily
// stripes updates across cache-line-padded cells selected via lock-free per-M randomness.
//
// Zero value is ready for immediate use. Do not copy a LongAdder after first use.
type LongAdder struct {
	striped64
}

// Add adds x to the counter.
func (l *LongAdder) Add(x int64) {
	tbl := l.cells.Load()
	if tbl == nil {
		b := atomic.LoadInt64(&l.base)
		if l.casBase(b, b+x) {
			return
		}
		l.longAccumulate(x, true)
		return
	}
	cs := tbl.cells
	m := uint32(len(cs) - 1)
	c := cs[rand.Uint32()&m]
	v := atomic.LoadInt64(&c.value)
	if !c.cas(v, v+x) {
		l.longAccumulate(x, false)
	}
}

// Inc increments the counter by 1.
func (l *LongAdder) Inc() {
	l.Add(1)
}

// Dec decrements the counter by 1.
func (l *LongAdder) Dec() {
	l.Add(-1)
}

// Sum returns the current sum across base and all allocated cells.
// Note: Sum is NOT an atomic snapshot; concurrent updates during Sum may or may not be included.
func (l *LongAdder) Sum() int64 {
	tbl := l.cells.Load()
	sum := atomic.LoadInt64(&l.base)
	if tbl != nil {
		for _, c := range tbl.cells {
			sum += atomic.LoadInt64(&c.value)
		}
	}
	return sum
}

// Reset zeroes the base and all allocated cells.
// Concurrent updates occurring simultaneously with Reset may be lost.
func (l *LongAdder) Reset() {
	tbl := l.cells.Load()
	atomic.StoreInt64(&l.base, 0)
	if tbl != nil {
		for _, c := range tbl.cells {
			c.reset(0)
		}
	}
}

// SumAndReset atomically swaps base and each allocated stripe to zero and returns the accumulated sum.
// It is not a linearizable point-in-time snapshot under concurrent updates, but because every stripe
// is drained with atomic Swap (while Add uses CAS), increments are conserved across successive calls.
func (l *LongAdder) SumAndReset() int64 {
	tbl := l.cells.Load()
	sum := atomic.SwapInt64(&l.base, 0)
	if tbl != nil {
		for _, c := range tbl.cells {
			sum += c.swap(0)
		}
	}
	return sum
}
