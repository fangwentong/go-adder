package adder_test

import (
	"context"
	"github.com/fangwentong/go-adder"
	"math/rand"
	"sync/atomic"
	"testing"
)

func BenchmarkLongAdder(b *testing.B) {
	var counter = &adder.LongAdder{}
	b.RunParallel(func(pb *testing.PB) {
		probe := rand.Int31()
		ctx := adder.ContextWithProbe(context.Background(), &probe)
		for pb.Next() {
			counter.Increment(ctx)
		}
	})
	if counter.Sum() != int64(b.N) {
		b.Fail()
	}
}

func BenchmarkLongAdderEmptyCtx(b *testing.B) {
	var counter = &adder.LongAdder{}
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			counter.Increment(ctx)
		}
	})
	if counter.Sum() != int64(b.N) {
		b.Fail()
	}
}

func BenchmarkAtomicLong(b *testing.B) {
	var val int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&val, 1)
		}
	})
	if atomic.LoadInt64(&val) != int64(b.N) {
		b.Fail()
	}
}
