package adder

import (
	"context"
	"sync/atomic"
)

type LongAdder struct {
	striped64
}

func (l *LongAdder) Add(x int64, ctx context.Context) {
	cs, _ := l.cells.Load().([]*cell)
	if cs == nil {
		b := atomic.LoadInt64(&l.base)
		if l.casBase(b, b+x) {
			return
		}
	}
	index := getProbe(ctx)
	if cs == nil {
		l.longAccumulate(x, nil, true, index, ctx)
		return
	}
	m := int32(len(cs)) - 1
	if m < 0 {
		l.longAccumulate(x, nil, true, index, ctx)
		return
	}
	c := cs[index&m]
	v := atomic.LoadInt64(&c.value)
	if !c.cas(v, v+x) {
		l.longAccumulate(x, nil, false, index, ctx)
	}
}

// Increment Equivalent to add(1)
func (l *LongAdder) Increment(ctx context.Context) {
	l.Add(1, ctx)
}

// Decrement Equivalent to add(-1)
func (l *LongAdder) Decrement(ctx context.Context) {
	l.Add(-1, ctx)
}

func (l *LongAdder) Sum() int64 {
	cs, _ := l.cells.Load().([]*cell)
	sum := atomic.LoadInt64(&l.base)
	if cs != nil {
		for _, c := range cs {
			sum += atomic.LoadInt64(&c.value)
		}
	}
	return sum
}

func (l *LongAdder) Reset() {
	cs, _ := l.cells.Load().([]*cell)
	atomic.StoreInt64(&l.base, 0)
	if cs != nil {
		for _, c := range cs {
			c.reset(0)
		}
	}
}

func (l *LongAdder) SumThenRest() int64 {
	cs, _ := l.cells.Load().([]*cell)
	sum := atomic.SwapInt64(&l.base, 0)
	if cs != nil {
		for _, c := range cs {
			sum += c.getAndSet(0)
		}
	}
	return sum
}
