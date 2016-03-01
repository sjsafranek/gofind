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
