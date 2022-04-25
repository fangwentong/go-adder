package adder

import (
	"context"
	"math/rand"
	"runtime"
	"sync/atomic"
)

type striped64 struct {
	cells     atomic.Value // []*cell
	base      int64        // volatile
	cellsBusy int32        // volatile
}

var nCPU = runtime.GOMAXPROCS(0)

func (s *striped64) casBase(cmp, val int64) bool {
	return atomic.CompareAndSwapInt64(&s.base, cmp, val)
}

func (s *striped64) getAndSetBase(val int64) int64 {
	return atomic.SwapInt64(&s.base, val)
}

func (s *striped64) casCellsBusy() bool {
	return atomic.CompareAndSwapInt32(&s.cellsBusy, 0, 1)
}

func (s *striped64) longAccumulate(ctx context.Context, x int64, fn func(int64, int64) int64, wasUncontended bool, index int32) {
	if index == 0 {
		index = initProbe(ctx)
		wasUncontended = true
	}
	collide := false
	for {
		cs, _ := s.cells.Load().([]*cell)
		n := len(cs)
		if cs != nil && n > 0 {
			c := cs[(n-1)&int(index)]
			if !wasUncontended { // CAS already known to fail
				wasUncontended = true // Continue after rehash
				index = advanceProbe(ctx, index)
				continue
			}
			v := atomic.LoadInt64(&c.value)
			var newVal int64
			if fn == nil {
				newVal = v + x
			} else {
				newVal = fn(v, x)
			}
			if c.cas(v, newVal) {
				break
			} else if n >= nCPU || !s.isSameReference(cs) {
				collide = false // At max size or stale
			} else if !collide {
				collide = true
			} else if atomic.LoadInt32(&s.cellsBusy) == 0 && s.casCellsBusy() {
				if s.isSameReference(cs) {
					rs := make([]*cell, n<<1)
					copy(rs, cs)
					for i := len(cs); i < len(rs); i++ {
						rs[i] = &cell{}
					}
					s.cells.Store(rs)
				}
				atomic.StoreInt32(&s.cellsBusy, 0)
				collide = false
				continue
			}
			index = advanceProbe(ctx, index)
		} else if atomic.LoadInt32(&s.cellsBusy) == 0 && s.isSameReference(cs) && s.casCellsBusy() {
			if s.isSameReference(cs) {
				rs := make([]*cell, 2)
				rs[index&1] = &cell{value: x}
				rs[1-index&1] = &cell{}
				s.cells.Store(rs)
				atomic.StoreInt32(&s.cellsBusy, 0)
				break
			}
			atomic.StoreInt32(&s.cellsBusy, 0)
		} else {
			v := atomic.LoadInt64(&s.base)
			var newVal int64
			if fn == nil {
				newVal = v + x
			} else {
				newVal = fn(v, x)
			}
			if s.casBase(v, newVal) {
				break
			}
		}
	}
}
func (s *striped64) isSameReference(test []*cell) bool {
	val, _ := s.cells.Load().([]*cell)
	if val == nil && test == nil {
		return true
	} else if val == nil || test == nil {
		return false
	}
	return &val[0] == &test[0]
}

var probeKey = &probeKeyType{}

type probeKeyType struct {
}

func ContextWithProbe(ctx context.Context, val *int32) context.Context {
	return context.WithValue(ctx, probeKey, val)
}

func initProbe(ctx context.Context) int32 {
	val, _ := ctx.Value(probeKey).(*int32)
	probe := rand.Int31()
	if val != nil {
		*val = probe
	}
	return probe
}

func getProbe(ctx context.Context) int32 {
	val, ok := ctx.Value(probeKey).(*int32)
	if val == nil || !ok {
		return 0
	} else {
		return *val
	}
}

func advanceProbe(ctx context.Context, probe int32) int32 {
	probe ^= probe << 13 // xorshift
	probe ^= probe >> 17
	probe ^= probe << 5

	val, _ := ctx.Value(probeKey).(*int32)
	if val != nil {
		*val = probe
	}
	return probe
}

type cell struct {
	value   int64    // volatile
	padding [7]int64 // padding for cpu cache-line align
}

func (c *cell) cas(cmp, val int64) bool {
	return atomic.CompareAndSwapInt64(&c.value, cmp, val)
}

func (c *cell) reset(identity int64) {
	atomic.StoreInt64(&c.value, identity)
}

func (c *cell) getAndSet(val int64) int64 {
	return atomic.SwapInt64(&c.value, val)
}
