package adder_test

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/fangwentong/go-adder"
)

func BenchmarkCounterComparison(b *testing.B) {
	parallelisms := []int{1, 8, runtime.GOMAXPROCS(0)}
	for _, p := range parallelisms {
		b.Run(fmt.Sprintf("LongAdder/p=%d", p), func(b *testing.B) {
			var counter adder.LongAdder
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					counter.Inc()
				}
			})
			if counter.Sum() != int64(b.N) {
				b.Fatalf("Sum() = %d, want %d", counter.Sum(), b.N)
			}
		})

		b.Run(fmt.Sprintf("AtomicInt64/p=%d", p), func(b *testing.B) {
			var counter atomic.Int64
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					counter.Add(1)
				}
			})
			if counter.Load() != int64(b.N) {
				b.Fatalf("Load() = %d, want %d", counter.Load(), b.N)
			}
		})
	}
}
