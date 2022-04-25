package adder_test

import (
	"context"
	"fmt"
	"github.com/fangwentong/go-adder"
	"sync"
	"sync/atomic"
	"testing"
)

func TestAddConcurrent(t *testing.T) {
	counter := &adder.LongAdder{}

	wg := sync.WaitGroup{}

	nInc := 50000000
	nRoutine := 20

	for i := 0; i < nRoutine; i++ {
		wg.Add(1)
		go func(i int) {
			var probe int32
			ctx := adder.ContextWithProbe(context.Background(), &probe)
			for j := 0; j < nInc; j++ {
				counter.Increment(ctx)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	sum := counter.Sum()

	if sum != int64(nInc)*int64(nRoutine) {
		t.Fail()
	}
	fmt.Println(sum)
}

func TestAddAtomic(t *testing.T) {
	var val int64

	wg := sync.WaitGroup{}

	nInc := 50000000
	nRoutine := 20

	for i := 0; i < nRoutine; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < nInc; j++ {
				atomic.AddInt64(&val, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	fmt.Println(val)
}
