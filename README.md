# gofind

[![Go Report Card](https://goreportcard.com/badge/github.com/schollz/gofind)](https://goreportcard.com/report/github.com/schollz/gofind)


# Benchmarks

Modern laptop
```
BenchmarkDumpFingerprint-4              2000        805391 ns/op
BenchmarkPutFingerprintInDatabase-4      200       8149515 ns/op
BenchmarkLoadFingerprint-4             30000         46877 ns/op
BenchmarkLoadParameters-4                100      15729660 ns/op
BenchmarkCalculatePosteriors1-4        30000         45821 ns/op
```

Raspberry Pi 2

```
BenchmarkDumpFingerprint-4         	     500	   3035738 ns/op
BenchmarkPutFingerprintInDatabase-4	       1	1079394176 ns/op
BenchmarkLoadFingerprint-4         	    2000	    579043 ns/op
BenchmarkLoadParameters-4          	       5	 228348392 ns/op
BenchmarkCalculatePosteriors1-4    	    1000	   1101443 ns/op
```
Not sure why `BenchmarkPutFingerprintInDatabase` is so slow in benchmarking, cause its not really, it took ~8-25 ms during actual trials.
