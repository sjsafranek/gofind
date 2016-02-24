package main

import "fmt"

import "encoding/json"

type Response2 struct {
	Group           string        `json:"group"`
	Username        string        `json:"username"`
	Location        string        `json:"location"`
	WifiFingerprint []Fingerprint `json:"wififingerprint"`
}

type Fingerprint struct {
	Mac  string `json:"mac"`
	Rssi int    `json:"rssi"`
}

var jsonExample = `{
	"group": "whatevergroup",
	"username": "iamauser",
	"location": null,
	"wififingerprint": [{
		"mac": "AA:AA:AA:AA:AA:AA",
		"rssi": -45
	}, {
		"mac": "BB:BB:BB:BB:BB:BB",
		"rssi": -55
	}]
}`

func main() {
	res := Response2{}
	json.Unmarshal([]byte(jsonExample), &res)
	fmt.Println(res)
	mapB, _ := json.Marshal(res)
	fmt.Println(string(mapB))
}
