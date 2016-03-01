package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

// PersistentParameters are not reloaded each time
type PersistentParameters struct {
	NetworkRenamed map[string][]string
}

// PriorParameters contains the network-specific bayesian priors and Mac frequency, as well as special variables
type PriorParameters struct {
	P        map[string]map[string][]float32 // standard P
	NP       map[string]map[string][]float32 // standard nP
	MacFreq  map[string]map[string]float32   // Frequency of a mac in a certain location
	NMacFreq map[string]map[string]float32   // Frequency of a mac, in everywhere BUT a certain location
	Special  map[string]float64
}

// ResultsParameters contains the information about the accuracy from crossValidation
type ResultsParameters struct {
	Accuracy         map[string]int            // accuracy measurement for a given location
	TotalLocations   map[string]int            // number of locations
	CorrectLocations map[string]int            // number of times guessed correctly
	Guess            map[string]map[string]int // correct -> guess -> times
}

// FullParameters is the full parameter set for a given group
type FullParameters struct {
	NetworkMacs    map[string]map[string]bool // map of networks and then the associated macs in each
	NetworkLocs    map[string]map[string]bool // map of the networks, and then the associated locations in each
	MacVariability map[string]float32         // variability of macs
	MacCount       map[string]int             // number of each mac
	MacCountByLoc  map[string]map[string]int  // number of each mac, by location
	UniqueLocs     []string
	UniqueMacs     []string
	Priors         map[string]PriorParameters   // generate priors for each network
	Results        map[string]ResultsParameters // generate priors for each network
	Loaded         bool                         // flag to determine if parameters have been loaded
}

// NewFullParameters generates a blank FullParameters
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
		Results:        make(map[string]ResultsParameters),
		Loaded:         false,
	}
}

// NewPriorParameters generates a blank PriorParameters
func NewPriorParameters() *PriorParameters {
	return &PriorParameters{
		P:        make(map[string]map[string][]float32),
		NP:       make(map[string]map[string][]float32),
		MacFreq:  make(map[string]map[string]float32),
		NMacFreq: make(map[string]map[string]float32),
		Special:  make(map[string]float64),
	}
}

// NewResultsParameters generates a blank ResultsParameters
func NewResultsParameters() *ResultsParameters {
	return &ResultsParameters{
		Accuracy:         make(map[string]int),
		TotalLocations:   make(map[string]int),
		CorrectLocations: make(map[string]int),
		Guess:            make(map[string]map[string]int),
	}
}

func NewPersistentParameters() *PersistentParameters {
	return &PersistentParameters{
		NetworkRenamed: make(map[string][]string),
	}
}

func dumpParameters(res FullParameters) []byte {
	defer timeTrack(time.Now(), "dumpParameters")
	jsonByte, _ := res.MarshalJSON()
	return compressByte(jsonByte)
}

func loadParameters(jsonByte []byte) FullParameters {
	var res2 FullParameters
	res2.UnmarshalJSON(decompressByte(jsonByte))
	return res2
}

func saveParameters(group string, res FullParameters) error {
	defer timeTrack(time.Now(), "saveParameters")
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("resources"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		err = bucket.Put([]byte("fullParameters"), dumpParameters(res))
		if err != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err
	})
	return err
}

func openParameters(group string) (FullParameters, error) {
	if _, ok := psCache[group]; ok {
		return psCache[group], nil
	}

	var ps = *NewFullParameters()
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("fullParameters"))
		ps = loadParameters(v)
		return nil
	})
	psCache[group] = ps
	return ps, err
}

func openPersistentParameters(group string) (PersistentParameters, error) {
	var persistentPs = *NewPersistentParameters()
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("resources"))
		if b == nil {
			return fmt.Errorf("Resources dont exist")
		}
		v := b.Get([]byte("persistentParameters"))
		json.Unmarshal(v, &persistentPs)
		return nil
	})
	return persistentPs, err
}

func savePersistentParameters(group string, res PersistentParameters) error {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("resources"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		jsonByte, _ := json.Marshal(res)
		err = bucket.Put([]byte("persistentParameters"), jsonByte)
		if err != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err
	})
	Debug.Println("Saved")
	return err
}

func getParameters(group string, ps *FullParameters) {
	defer timeTrack(time.Now(), "getParameters")
	persistentPs, err := openPersistentParameters(group)
	ps.NetworkMacs = make(map[string]map[string]bool)
	ps.NetworkLocs = make(map[string]map[string]bool)
	ps.UniqueMacs = []string{}
	ps.UniqueLocs = []string{}
	ps.MacCount = make(map[string]int)
	ps.MacCountByLoc = make(map[string]map[string]int)
	ps.Loaded = true
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

	// Rename the NetworkMacs
	if len(persistentPs.NetworkRenamed) > 0 {
		fmt.Println("Figuring out the renaming...")
		newNames := []string{}
		for k := range persistentPs.NetworkRenamed {
			newNames = append(newNames, k)
		}
		for n := range ps.NetworkMacs {
			renamed := false
			for mac := range ps.NetworkMacs[n] {
				for renamedN := range persistentPs.NetworkRenamed {
					if stringInSlice(mac, persistentPs.NetworkRenamed[renamedN]) && !stringInSlice(n, newNames) {
						ps.NetworkMacs[renamedN] = make(map[string]bool)
						for k, v := range ps.NetworkMacs[n] {
							ps.NetworkMacs[renamedN][k] = v
						}
						delete(ps.NetworkMacs, n)
						renamed = true
					}
					if renamed {
						break
					}
				}
				if renamed {
					break
				}
			}
		}
	}

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
