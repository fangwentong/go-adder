# go-adder

A high-concurrency striped `int64` counter for Go (Go 1.22+), inspired by Java's `LongAdder`.

Under low contention it performs a single atomic CAS on `base`. Under high contention it lazily stripes updates across
128-byte cache-line-padded cells selected using Go 1.22+'s lock-free per-M `math/rand/v2` without requiring `context`
plumbing or `go:linkname`.

## Usage

```go
package main

import (
	"fmt"

	"github.com/fangwentong/go-adder"
)

func main() {
	var counter adder.LongAdder // zero value is ready to use

	counter.Add(2)
	counter.Inc()
	counter.Dec()

	fmt.Println(counter.Sum())         // 2
	fmt.Println(counter.SumAndReset()) // 2 (then zeroed)
	counter.Reset()
}
```

## Consistency Semantics

- `Sum()` iterates over `base` and allocated stripes; it is **not** an atomic snapshot under concurrent updates.
- `SumAndReset()` is also not a linearizable point-in-time snapshot, but because it drains `base` and every stripe
  with atomic `Swap`, increments are conserved across successive calls without loss.
- `Reset()` uses plain atomic stores; concurrent updates racing with `Reset()` may be lost.
- Do not copy a `LongAdder` after first use.

## Benchmark

Measured with `go test -bench=BenchmarkCounterComparison -benchtime=300ms` on `darwin/arm64` (Apple M5 Max,
`GOMAXPROCS=18`, where `p` is `b.SetParallelism(p)` = `p * GOMAXPROCS` goroutines):

| Parallelism (`p`) | `LongAdder.Inc()` | `atomic.Int64.Add(1)` |
| --- | --- | --- |
| `p=1` (18 goroutines) | `12.91 ns/op` | `13.51 ns/op` |
| `p=8` (144 goroutines) | `12.33 ns/op` | `14.55 ns/op` |
| `p=18` (324 goroutines) | `11.90 ns/op` | `14.47 ns/op` |

## License

MIT License. See [LICENSE](LICENSE).
