go-adder
---

This is a port of
the [Java's LongAdder](https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/util/concurrent/atomic/LongAdder.html)
into the Go programming language.

## Usage

```go
package test

import "github.com/fangwentong/go-adder"
import "context"

var counter = &adder.LongAdder{}

func ExampleFunc(ctx context.Context)  {
    // inject goroutine probe
    var probe int32
    ctx = adder.ContextWithProbe(ctx, &probe)

    counter.Add(ctx, 2)
    counter.Add(ctx, -2)

    // counter +1
    counter.Increment(ctx)
    // counter -1
    counter.Decrement(ctx)

    // get current count
    counter.Sum()

    // reset current count
    counter.Reset()

    // get current count then reset
    counter.SumThenRest()
}
```

## Benchmark

running on MacOS powered by Intel [Coffee Lake](https://en.wikichip.org/wiki/intel/microarchitectures/coffee_lake) 6
core 12 thread

```
$ go test -run=nothing -bench=. -cpu=1,2,4,8,16,32,64,128 -benchtime=5s
goos: darwin
goarch: amd64
pkg: github.com/fangwentong/go-adder
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkLongAdder                      691230033                8.824 ns/op
BenchmarkLongAdder-2                    687713216               18.45 ns/op
BenchmarkLongAdder-4                    477778184               12.29 ns/op
BenchmarkLongAdder-8                    1000000000               2.966 ns/op
BenchmarkLongAdder-16                   1000000000               2.573 ns/op
BenchmarkLongAdder-32                   1000000000               2.025 ns/op
BenchmarkLongAdder-64                   1000000000               2.373 ns/op
BenchmarkLongAdder-128                  1000000000               2.297 ns/op
BenchmarkLongAdderEmptyCtx              519916878               12.42 ns/op
BenchmarkLongAdderEmptyCtx-2            160806240               36.50 ns/op
BenchmarkLongAdderEmptyCtx-4            172157611               32.71 ns/op
BenchmarkLongAdderEmptyCtx-8            222932665               28.26 ns/op
BenchmarkLongAdderEmptyCtx-16           231772944               26.73 ns/op
BenchmarkLongAdderEmptyCtx-32           238532882               24.33 ns/op
BenchmarkLongAdderEmptyCtx-64           254754256               24.61 ns/op
BenchmarkLongAdderEmptyCtx-128          244789734               24.35 ns/op
BenchmarkAtomicLong                     802940997                6.566 ns/op
BenchmarkAtomicLong-2                   366768531               15.57 ns/op
BenchmarkAtomicLong-4                   388438527               15.96 ns/op
BenchmarkAtomicLong-8                   361224614               16.80 ns/op
BenchmarkAtomicLong-16                  385923369               18.04 ns/op
BenchmarkAtomicLong-32                  370266336               15.54 ns/op
BenchmarkAtomicLong-64                  387258997               16.38 ns/op
BenchmarkAtomicLong-128                 307278933               18.82 ns/op
PASS
ok      github.com/fangwentong/go-adder 177.124s
```

