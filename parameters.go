package main

import (
	"encoding/json"
	"log"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

type PriorParameters struct {
	Usefulness float64
	Mixin      float64
	P          map[string]map[string][]float32 // standard P
	NP         map[string]map[string][]float32 // standard nP
	MacFreq    map[string]map[string]float32   // Frequency of a mac in a certain location
	NMacFreq   map[string]map[string]float32   // Frequency of a mac, in everywhere BUT a certain location
}

// Array of parameters used for the network
type FullParameters struct {
	NetworkMacs    map[string]map[string]bool // map of networks and then the associated macs in each
	NetworkLocs    map[string]map[string]bool // map of the networks, and then the associated locations in each
	MacVariability map[string]float32         // variability of macs
	MacCount       map[string]int             // number of each mac
	MacCountByLoc  map[string]map[string]int  // number of each mac, by location
	UniqueLocs     []string
	UniqueMacs     []string
	Priors         map[string]PriorParameters // generate priors for each network
}

func NewFullParameters() *FullParameters {
	return &FullParameters{
		NetworkMacs:    make(map[string]map[string]bool),
		NetworkLocs:    make(map[string]map[string]bool),
		MacCount:       make(map[string]int),
		MacCountByLoc:  make(map[string]map[string]int),
		UniqueMacs:     []string{},
		UniqueLocs:     []string{},
		Priors:         make(map[string]PriorParameters),
		MacVariability: make(map[string]float32),
	}
}

func NewPriorParameters() *PriorParameters {
	return &PriorParameters{
		Usefulness: 0,
		Mixin:      0.5,
		P:          make(map[string]map[string][]float32),
		NP:         make(map[string]map[string][]float32),
		MacFreq:    make(map[string]map[string]float32),
		NMacFreq:   make(map[string]map[string]float32),
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

func getParameters(group string, ps *FullParameters) {
	defer timeTrack(time.Now(), "getParameters")
	ps.NetworkMacs = make(map[string]map[string]bool)
	ps.NetworkLocs = make(map[string]map[string]bool)
	ps.UniqueMacs = []string{}
	ps.UniqueLocs = []string{}
	ps.MacCount = make(map[string]int)
	ps.MacCountByLoc = make(map[string]map[string]int)
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get all parameters that don't need a network graph
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)

			// unique locs
			if !stringInSlice(v2.Location, ps.UniqueLocs) {
				ps.UniqueLocs = append(ps.UniqueLocs, v2.Location)
			}

			// mac by location count
			if _, ok := ps.MacCountByLoc[v2.Location]; !ok {
				ps.MacCountByLoc[v2.Location] = make(map[string]int)
			}

			// building network
			macs := []string{}

			for _, router := range v2.WifiFingerprint {
				// building network
				macs = append(macs, router.Mac)

				// unique macs
				if !stringInSlice(router.Mac, ps.UniqueMacs) {
					ps.UniqueMacs = append(ps.UniqueMacs, router.Mac)
				}

				// mac count
				if _, ok := ps.MacCount[router.Mac]; !ok {
					ps.MacCount[router.Mac] = 0
				}
				ps.MacCount[router.Mac]++

				// mac by location count
				if _, ok := ps.MacCountByLoc[v2.Location][router.Mac]; !ok {
					ps.MacCountByLoc[v2.Location][router.Mac] = 0
				}
				ps.MacCountByLoc[v2.Location][router.Mac]++
			}

			// building network
			ps.NetworkMacs = buildNetwork(ps.NetworkMacs, macs)
		}
		return nil
	})
	ps.NetworkMacs = mergeNetwork(ps.NetworkMacs)

	// Get the locations for each graph (Has to have network built first)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			macs := []string{}
			for _, router := range v2.WifiFingerprint {
				macs = append(macs, router.Mac)
			}
			networkName, inNetwork := hasNetwork(ps.NetworkMacs, macs)
			if inNetwork {
				if _, ok := ps.NetworkLocs[networkName]; !ok {
					ps.NetworkLocs[networkName] = make(map[string]bool)
				}
				if _, ok := ps.NetworkLocs[networkName][v2.Location]; !ok {
					ps.NetworkLocs[networkName][v2.Location] = true
				}
			}
		}
		return nil
	})

}
