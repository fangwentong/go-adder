package adder_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/fangwentong/go-adder"
)

func TestAddConcurrent(t *testing.T) {
	var counter adder.LongAdder

	nInc := 100_000
	if testing.Short() {
		nInc = 20_000
	}
	nRoutine := 16

	var wg sync.WaitGroup
	for i := 0; i < nRoutine; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < nInc; j++ {
				if id%2 == 0 {
					counter.Inc()
				} else {
					counter.Add(2)
					counter.Dec()
				}
			}
		}(i)
	}
	wg.Wait()

	want := int64(nInc) * int64(nRoutine)
	if got := counter.Sum(); got != want {
		t.Fatalf("Sum() = %d, want %d", got, want)
	}
}

func TestResetAndSumAndReset(t *testing.T) {
	var counter adder.LongAdder
	counter.Add(10)
	counter.Dec()
	if got := counter.Sum(); got != 9 {
		t.Fatalf("Sum() = %d, want 9", got)
	}
	if got := counter.SumAndReset(); got != 9 {
		t.Fatalf("SumAndReset() = %d, want 9", got)
	}
	if got := counter.Sum(); got != 0 {
		t.Fatalf("Sum() after SumAndReset() = %d, want 0", got)
	}

	counter.Add(-5)
	if got := counter.Sum(); got != -5 {
		t.Fatalf("Sum() with negative delta = %d, want -5", got)
	}
	counter.Reset()
	if got := counter.Sum(); got != 0 {
		t.Fatalf("Sum() after Reset() = %d, want 0", got)
	}
}

func TestConcurrentMixedAddSumReset(t *testing.T) {
	var counter adder.LongAdder
	var drained atomic.Int64

	const writers = 8
	const perWriter = 10_000
	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWriter; j++ {
				counter.Inc()
				if j%500 == 0 {
					_ = counter.Sum()
				}
			}
		}()
	}

	done := make(chan struct{})
	var drainerWG sync.WaitGroup
	drainerWG.Add(1)
	go func() {
		defer drainerWG.Done()
		for {
			select {
			case <-done:
				return
			default:
				drained.Add(counter.SumAndReset())
			}
		}
	}()

	wg.Wait()
	close(done)
	drainerWG.Wait()
	drained.Add(counter.SumAndReset())

	// SumAndReset drains base and each stripe with atomic Swap (while Add uses CAS),
	// and table expansion preserves existing cell pointers, so every increment must be
	// conserved across all SumAndReset calls once writers finish.
	const want = int64(writers * perWriter)
	if got := drained.Load(); got != want {
		t.Fatalf("drained total = %d, want %d", got, want)
	}
}
