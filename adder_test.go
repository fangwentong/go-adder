package adder_test

import (
	"context"
	"sync"
	"testing"

	"github.com/fangwentong/go-adder"
)

func TestAddConcurrent(t *testing.T) {
	counter := &adder.LongAdder{}

	nInc := 100_000
	if testing.Short() {
		nInc = 20_000
	}
	nRoutine := 16

	var wg sync.WaitGroup
	for i := 0; i < nRoutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var probe int32
			ctx := adder.ContextWithProbe(context.Background(), &probe)
			for j := 0; j < nInc; j++ {
				counter.Increment(ctx)
			}
		}()
	}
	wg.Wait()

	want := int64(nInc) * int64(nRoutine)
	if got := counter.Sum(); got != want {
		t.Fatalf("Sum() = %d, want %d", got, want)
	}
}

func TestResetAndSumThenRest(t *testing.T) {
	counter := &adder.LongAdder{}
	var probe int32
	ctx := adder.ContextWithProbe(context.Background(), &probe)

	counter.Add(ctx, 10)
	counter.Decrement(ctx)
	if got := counter.Sum(); got != 9 {
		t.Fatalf("Sum() = %d, want 9", got)
	}

	if got := counter.SumThenRest(); got != 9 {
		t.Fatalf("SumThenRest() = %d, want 9", got)
	}
	if got := counter.Sum(); got != 0 {
		t.Fatalf("Sum() after SumThenRest() = %d, want 0", got)
	}

	counter.Add(ctx, 42)
	counter.Reset()
	if got := counter.Sum(); got != 0 {
		t.Fatalf("Sum() after Reset() = %d, want 0", got)
	}
}
