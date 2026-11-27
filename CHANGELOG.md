# Changelog

## 0.2.0

### Breaking Changes
- Minimum Go version bumped to **Go 1.22** (for lock-free `math/rand/v2`).
- Simplified public API by removing `context.Context` / `ContextWithProbe`:
  - `Add(x int64)`, `Inc()`, `Dec()`, `Sum() int64`, `Reset()`, `SumAndReset() int64`.
- Fixed method typo `SumThenRest` → `SumAndReset`.

### Fixes & Improvements
- Eliminated data race when sharing a `context.Context` across goroutines and eliminated `int32` arithmetic right-shift
  probe collapse onto `cell[0]`.
- Replaced `atomic.Value` cell slice storage with `atomic.Pointer[cellTable]` (zero interface boxing, pointer-identity
  CAS checks).
- Padded `cell` and `striped64` fields to 128 bytes to avoid false sharing on ARM64 / Apple Silicon as well as x86_64.
- Bounded stripe table size from `runtime.GOMAXPROCS(0)` (cached on table creation/resize so the hot contention loop never acquires the Go scheduler lock).
- Added mixed concurrent `Add`/`Sum`/`SumAndReset` `-race` tests and comparison benchmarks against `atomic.Int64`.
