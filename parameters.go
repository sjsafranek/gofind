package main

import "encoding/json"

type PriorParameters struct {
	NetworkName string
	Usefulness  float64
	Mixin       float64
	P           map[string]map[string][]float32 // standard P
	nP          map[string]map[string][]float32 // standard nP
	MacFreq     map[string]map[string]float32   // Frequency of a mac in a certain location
	nMacFreq    map[string]map[string]float32   // Frequency of a mac, in everywhere BUT a certain location
}

// Array of parameters used for the network
type FullParameters struct {
	NetworkPriors []PriorParameters          // generate priors for each network
	NetworkMacs   map[string]map[string]bool // map of networks and then the associated macs in each
	NetworkLocs   map[string]map[string]bool // map of the networks, and then the associated locations in each
	MacCount      map[string]int             // number of each mac
	MacCountByLoc map[string]map[string]int  // number of each mac, by location
	UsefulMacs    map[string]bool            // list of useful macs
}

func NewFullParameters() *FullParameters {
	return &FullParameters{
		NetworkPriors: []PriorParameters{},
		NetworkMacs:   make(map[string]map[string]bool),
		NetworkLocs:   make(map[string]map[string]bool),
		MacCount:      make(map[string]int),
		MacCountByLoc: make(map[string]map[string]int),
		UsefulMacs:    make(map[string]bool),
	}
}

func NewPriorParameters() *PriorParameters {
	return &PriorParameters{
		NetworkName: "-1",
		Usefulness:  0,
		Mixin:       0.5,
		P:           make(map[string]map[string][]float32),
		nP:          make(map[string]map[string][]float32),
		MacFreq:     make(map[string]map[string]float32),
		nMacFreq:    make(map[string]map[string]float32),
	}
}

func dumpParameters(res FullParameters) []byte {
	jsonByte, _ := json.Marshal(res)
	return jsonByte
}

func loadParameters(jsonByte []byte) FullParameters {
	var res2 FullParameters
	json.Unmarshal(jsonByte, &res2)
	return res2
}
