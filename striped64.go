package adder

import (
	"math/rand/v2"
	"runtime"
	"sync/atomic"
)

const cacheLineSize = 128

// cell holds a single striped int64 counter padded to 128 bytes to avoid false sharing
// on both 64-byte (x86_64) and 128-byte (Apple Silicon / ARM64) cache-line architectures.
type cell struct {
	value int64
	_     [cacheLineSize - 8]byte
}

func (c *cell) cas(cmp, val int64) bool {
	return atomic.CompareAndSwapInt64(&c.value, cmp, val)
}

func (c *cell) reset(val int64) {
	atomic.StoreInt64(&c.value, val)
}

func (c *cell) swap(val int64) int64 {
	return atomic.SwapInt64(&c.value, val)
}

// cellTable wraps the striped cell slice so atomic.Pointer can swap tables without
// interface boxing allocations and compare table identity via pointer equality.
// maxLen caches the target stripe cap at table creation/resize so the hot contention
// path never calls runtime.GOMAXPROCS(0) (which acquires the global scheduler lock).
type cellTable struct {
	cells  []*cell
	maxLen int
}

type striped64 struct {
	cells     atomic.Pointer[cellTable]
	_         [cacheLineSize - 8]byte
	base      int64
	_         [cacheLineSize - 8]byte
	cellsBusy int32
	_         [cacheLineSize - 4]byte
}

func (s *striped64) casBase(cmp, val int64) bool {
	return atomic.CompareAndSwapInt64(&s.base, cmp, val)
}

func (s *striped64) casCellsBusy() bool {
	return atomic.CompareAndSwapInt32(&s.cellsBusy, 0, 1)
}

func maxCells() int {
	p := runtime.GOMAXPROCS(0)
	if p <= 2 {
		return 2
	}
	// Round up to the next power of two so index masking (len-1) stays uniform.
	n := 1
	for n < p {
		n <<= 1
	}
	return n
}

func (s *striped64) longAccumulate(x int64, wasUncontended bool) {
	probe := rand.Uint32()
	collide := false
	for {
		tbl := s.cells.Load()
		if tbl != nil && len(tbl.cells) > 0 {
			cs := tbl.cells
			n := len(cs)
			c := cs[int(probe)&(n-1)]
			if !wasUncontended {
				wasUncontended = true
				probe = rand.Uint32()
				continue
			}
			v := atomic.LoadInt64(&c.value)
			if c.cas(v, v+x) {
				return
			}
			if n >= tbl.maxLen || s.cells.Load() != tbl {
				collide = false
			} else if !collide {
				collide = true
			} else if atomic.LoadInt32(&s.cellsBusy) == 0 && s.casCellsBusy() {
				if s.cells.Load() == tbl {
					capLimit := maxCells()
					if n < capLimit {
						rs := make([]*cell, n<<1)
						copy(rs, cs)
						for i := n; i < len(rs); i++ {
							rs[i] = &cell{}
						}
						s.cells.Store(&cellTable{cells: rs, maxLen: capLimit})
					} else {
						// Refresh cached cap if GOMAXPROCS decreased since table allocation.
						s.cells.Store(&cellTable{cells: cs, maxLen: capLimit})
					}
				}
				atomic.StoreInt32(&s.cellsBusy, 0)
				collide = false
				continue
			}
			probe = rand.Uint32()
		} else if atomic.LoadInt32(&s.cellsBusy) == 0 && s.cells.Load() == tbl && s.casCellsBusy() {
			if s.cells.Load() == tbl {
				rs := []*cell{{}, {}}
				rs[probe&1].value = x
				s.cells.Store(&cellTable{cells: rs, maxLen: maxCells()})
				atomic.StoreInt32(&s.cellsBusy, 0)
				return
			}
			atomic.StoreInt32(&s.cellsBusy, 0)
		} else {
			v := atomic.LoadInt64(&s.base)
			if s.casBase(v, v+x) {
				return
			}
		}
	}
}
