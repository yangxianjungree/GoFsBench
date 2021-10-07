# COMMON

## 1. 日志性能测试

```s
# go test -bench=. -cpu=16 -benchtime="5s" -benchmem

goos: linux
goarch: amd64
pkg: common
cpu: Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz
BenchmarkLogging-16      2915629          2065 ns/op          56 B/op          2 allocs/op
PASS
ok      common  8.108s
```
