package main

var PdfType []float64
var MaxRssi int
var MinRssi int
var RssiPartitions int

func init() {
	PdfType = []float64{.1995, .1760, .1210, .0648, .027, 0.005}
	MinRssi = -100
	MaxRssi = -10
	RssiPartitions = MaxRssi - MinRssi + 1
}

func calculatePriors(group string) {
}
