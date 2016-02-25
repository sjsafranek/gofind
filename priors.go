package main

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

var PdfType []float32
var MaxRssi int
var MinRssi int
var RssiPartitions int
var Absentee float32

func init() {
	PdfType = []float32{.1995, .1760, .1210, .0648, .027, 0.005}
	Absentee = 1e-6
	MinRssi = -100
	MaxRssi = -10
	RssiPartitions = MaxRssi - MinRssi + 1
}

func calculatePriors(group string, ps *FullParameters) {
	defer timeTrack(time.Now(), "calculatePriors")
	ps.Priors = make(map[string]PriorParameters)
	for n := range ps.NetworkLocs {
		var newPrior = *NewPriorParameters()
		ps.Priors[n] = newPrior
	}

	// Initialization
	for n := range ps.Priors {
		for loc := range ps.NetworkLocs[n] {
			ps.Priors[n].P[loc] = make(map[string][]float32)
			ps.Priors[n].NP[loc] = make(map[string][]float32)
			for mac := range ps.NetworkMacs[n] {
				ps.Priors[n].P[loc][mac] = make([]float32, RssiPartitions)
				ps.Priors[n].NP[loc][mac] = make([]float32, RssiPartitions)
			}
		}
	}

	// Add in Priors from database
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
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
				for _, router := range v2.WifiFingerprint {
					ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi] += PdfType[0]
					for i, val := range PdfType {
						if i > 0 {
							ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi-i] += val
							ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi+i] += val
						}
					}
				}
			}

		}
		return nil
	})
	db.Close()

	// Calculate the nP
	for n := range ps.Priors {
		for locN := range ps.NetworkLocs[n] {
			for loc := range ps.NetworkLocs[n] {
				if loc != locN {
					for mac := range ps.NetworkMacs[n] {
						for i := range ps.Priors[n].P[locN][mac] {
							(ps.Priors[n].NP[locN][mac][i]) += ps.Priors[n].P[loc][mac][i]
						}
					}
				}
			}
		}
	}

	// normalize P and nP
	rssiRange := make([]float32, RssiPartitions)
	for i := 0; i < len(rssiRange); i++ {
		rssiRange[i] = float32(MinRssi + i)
	}
	for n := range ps.Priors {
		macAverages := make(map[string][]float32)
		for loc := range ps.NetworkLocs[n] {
			for mac := range ps.NetworkMacs[n] {
				total := float32(0)
				for _, val := range ps.Priors[n].P[loc][mac] {
					total += val
				}
				averageMac := float32(0)
				for i, val := range ps.Priors[n].P[loc][mac] {
					if val > float32(0) {
						ps.Priors[n].P[loc][mac][i] = val / total
						averageMac += rssiRange[i] * ps.Priors[n].P[loc][mac][i]
					}
				}
				if averageMac < float32(0) {
					if _, ok := macAverages[mac]; !ok {
						macAverages[mac] = []float32{}
					}
					macAverages[mac] = append(macAverages[mac], averageMac)
				}

				total = float32(0)
				for i := range ps.Priors[n].NP[loc][mac] {
					total += ps.Priors[n].NP[loc][mac][i]
				}
				for i := range ps.Priors[n].NP[loc][mac] {
					ps.Priors[n].NP[loc][mac][i] = ps.Priors[n].NP[loc][mac][i] / total
				}
			}
		}
		fmt.Println(macAverages)
	}

}
