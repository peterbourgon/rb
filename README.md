# rb [![Go Reference](https://pkg.go.dev/badge/github.com/peterbourgon/rb.svg)](https://pkg.go.dev/github.com/peterbourgon/rb) [![GitHub Release](https://img.shields.io/github/v/release/peterbourgon/rb?style=flat-square)](https://github.com/peterbourgon/rb/releases) [![Build Status](https://github.com/peterbourgon/rb/actions/workflows/test.yaml/badge.svg)](https://github.com/peterbourgon/rb/actions/workflows/test.yaml)

In-memory ring buffer optimized for high-throughput write-heavy workloads.

Buffers are fully allocated during construction, so adds are zero-alloc and reasonably fast.

```
goos: darwin
goarch: arm64
pkg: github.com/peterbourgon/rb
cpu: Apple M3 Pro
BenchmarkRingBuffer/cap=100/Add-11      267911362   4.408 ns/op   0 B/op   0 allocs/op
BenchmarkRingBuffer/cap=1000/Add-11     270116073   4.446 ns/op   0 B/op   0 allocs/op
BenchmarkRingBuffer/cap=10000/Add-11    267690060   4.456 ns/op   0 B/op   0 allocs/op
BenchmarkRingBuffer/cap=100000/Add-11   269908416   4.450 ns/op   0 B/op   0 allocs/op
BenchmarkRingBuffer/cap=1000000/Add-11  266172448   4.563 ns/op   0 B/op   0 allocs/op
```

It's pretty self-explanatory. [See the documentation](https://pkg.go.dev/github.com/peterbourgon/rb) for details.
