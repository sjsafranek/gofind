package main

import "encoding/json"

type Parameters struct {
	Mixin      float64
	Usefulness float64
}

func dumpParameters(res Parameters) []byte {
	jsonByte, _ := json.Marshal(res)
	return jsonByte
}

func loadParameters(jsonByte []byte) Parameters {
	var res2 Parameters
	json.Unmarshal(jsonByte, &res2)
	return res2
}
