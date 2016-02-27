package main

import (
	"fmt"
	"log"
	"math"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

var PdfType []float32
var MaxRssi int
var MinRssi int
var RssiPartitions int
var Absentee float32
var RssiRange []float32

func init() {
	PdfType = []float32{.1995, .1760, .1210, .0648, .027, 0.005}
	Absentee = 1e-6
	MinRssi = -100
	MaxRssi = -10
	RssiPartitions = MaxRssi - MinRssi + 1
	RssiRange = make([]float32, RssiPartitions)
	for i := 0; i < len(RssiRange); i++ {
		RssiRange[i] = float32(MinRssi + i)
	}
}

func optimizePriors(group string) {
	defer timeTrack(time.Now(), "optimizePriors")
	var ps FullParameters = *NewFullParameters()
	getParameters(group, &ps)
	calculatePriors(group, &ps)
	// fmt.Println(string(dumpParameters(ps)))
	// ps, _ = openParameters("findtest")
	var results ResultsParameters = *NewResultsParameters()
	for n := range ps.Priors {
		ps.Results[n] = results
	}
	// fmt.Println(ps.Results)
	// ps.Priors["0"].Special["MixIn"] = 1.0
	// fmt.Println(crossValidation(group, "0", &ps))
	// fmt.Println(ps.Results)

	mixins := []float64{0.1, 0.25, 0.5, 0.75, 0.9}
	cutoffs := []float64{0.05, 0.1, 0.15, 0.2, 0.5}

	for n := range ps.Priors {
		bestResult := float64(0)
		bestMixin := float64(0)
		bestCutoff := float64(0)
		for _, mixin := range mixins {
			for _, cutoff := range cutoffs {
				ps.Priors[n].Special["MixIn"] = mixin
				ps.Priors[n].Special["VarabilityCutoff"] = cutoff
				avgAccuracy := crossValidation(group, n, &ps)
				if avgAccuracy > bestResult {
					bestResult = avgAccuracy
					bestMixin = mixin
					bestCutoff = cutoff
				}
			}
		}
		ps.Priors[n].Special["MixIn"] = bestMixin
		ps.Priors[n].Special["VarabilityCutoff"] = bestCutoff
		fmt.Println(bestMixin, bestResult)
		Debug.Println("Final cross validation...")
		crossValidation(group, n, &ps)
	}

	saveParameters(group, ps)
}

func regenerateEverything(group string) {
	defer timeTrack(time.Now(), "regenerateParameters")
	var ps FullParameters = *NewFullParameters()
	ps, _ = openParameters(group)
	getParameters(group, &ps)
	calculatePriors(group, &ps)
	var results ResultsParameters = *NewResultsParameters()
	for n := range ps.Priors {
		ps.Results[n] = results
	}
	for n := range ps.Priors {
		crossValidation(group, n, &ps)
	}
	saveParameters(group, ps)
}

func crossValidation(group string, n string, ps *FullParameters) float64 {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for loc := range ps.NetworkLocs[n] {
		ps.Results[n].TotalLocations[loc] = 0
		ps.Results[n].CorrectLocations[loc] = 0
		ps.Results[n].Accuracy[loc] = 0
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		it := float64(0)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if math.Mod(it, 2) == 0 {
				v2 := loadFingerprint(v)
				if _, ok := ps.NetworkLocs[n][v2.Location]; ok {
					locationGuess, _ := calculatePosterior(v2, *ps)
					ps.Results[n].TotalLocations[v2.Location]++
					if locationGuess == v2.Location {
						ps.Results[n].CorrectLocations[v2.Location]++
					}
					if _, ok := ps.Results[n].Guess[v2.Location]; !ok {
						ps.Results[n].Guess[v2.Location] = make(map[string]int)
					}
					if _, ok := ps.Results[n].Guess[v2.Location][locationGuess]; !ok {
						ps.Results[n].Guess[v2.Location][locationGuess] = 0
					}
					ps.Results[n].Guess[v2.Location][locationGuess]++
				}
			}
			it++
		}
		return nil
	})

	average := float64(0)
	for loc := range ps.NetworkLocs[n] {
		if ps.Results[n].TotalLocations[loc] > 0 {
			ps.Results[n].Accuracy[loc] = int(100.0 * ps.Results[n].CorrectLocations[loc] / ps.Results[n].TotalLocations[loc])
			average += float64(ps.Results[n].Accuracy[loc])
		}
	}
	average = average / float64(len(ps.NetworkLocs[n]))

	return average
}

func calculatePriors(group string, ps *FullParameters) {
	defer timeTrack(time.Now(), "calculatePriors")
	ps.Priors = make(map[string]PriorParameters)
	for n := range ps.NetworkLocs {
		var newPrior = *NewPriorParameters()
		ps.Priors[n] = newPrior
	}

	// Initialization
	ps.MacVariability = make(map[string]float32)
	for n := range ps.Priors {
		ps.Priors[n].Special["MacFreqMin"] = float64(100)
		ps.Priors[n].Special["NMacFreqMin"] = float64(100)
		for loc := range ps.NetworkLocs[n] {
			ps.Priors[n].P[loc] = make(map[string][]float32)
			ps.Priors[n].NP[loc] = make(map[string][]float32)
			ps.Priors[n].MacFreq[loc] = make(map[string]float32)
			ps.Priors[n].NMacFreq[loc] = make(map[string]float32)
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
		it := float64(-1)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			it++
			if math.Mod(it, 2) == 0 { // cross-validation
				continue
			}
			v2 := loadFingerprint(v)
			macs := []string{}
			for _, router := range v2.WifiFingerprint {
				macs = append(macs, router.Mac)
			}

			networkName, inNetwork := hasNetwork(ps.NetworkMacs, macs)
			if inNetwork {
				for _, router := range v2.WifiFingerprint {
					if router.Rssi > MinRssi {
						ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi] += PdfType[0]
						for i, val := range PdfType {
							if i > 0 {
								ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi-i] += val
								ps.Priors[networkName].P[v2.Location][router.Mac][router.Rssi-MinRssi+i] += val
							}
						}
					} else {
						Warning.Println(router.Rssi)
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
							if ps.Priors[n].P[loc][mac][i] > 0 {
								ps.Priors[n].NP[locN][mac][i] += ps.Priors[n].P[loc][mac][i]
							}
						}
					}
				}
			}
		}
	}

	// Add in absentee, normalize P and nP and determine MacVariability
	for n := range ps.Priors {
		macAverages := make(map[string][]float32)

		for loc := range ps.NetworkLocs[n] {
			for mac := range ps.NetworkMacs[n] {
				for i := range ps.Priors[n].P[loc][mac] {
					ps.Priors[n].P[loc][mac][i] += Absentee
					ps.Priors[n].NP[loc][mac][i] += Absentee
				}
				total := float32(0)
				for _, val := range ps.Priors[n].P[loc][mac] {
					total += val
				}
				averageMac := float32(0)
				for i, val := range ps.Priors[n].P[loc][mac] {
					if val > float32(0) {
						ps.Priors[n].P[loc][mac][i] = val / total
						averageMac += RssiRange[i] * ps.Priors[n].P[loc][mac][i]
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
				if total > 0 {
					for i := range ps.Priors[n].NP[loc][mac] {
						ps.Priors[n].NP[loc][mac][i] = ps.Priors[n].NP[loc][mac][i] / total
					}
				}
			}
		}

		// Determine MacVariability
		for mac := range macAverages {
			if len(macAverages[mac]) <= 2 {
				ps.MacVariability[mac] = float32(1)
			} else {
				maxVal := float32(-10000)
				for _, val := range macAverages[mac] {
					if val > maxVal {
						maxVal = val
					}
				}
				for i, val := range macAverages[mac] {
					macAverages[mac][i] = maxVal / val
				}
				ps.MacVariability[mac] = standardDeviation(macAverages[mac])
			}
		}
	}

	// Determine mac frequencies and normalize
	for n := range ps.Priors {
		for loc := range ps.NetworkLocs[n] {
			maxCount := 0
			for mac := range ps.MacCountByLoc[loc] {
				if ps.MacCountByLoc[loc][mac] > maxCount {
					maxCount = ps.MacCountByLoc[loc][mac]
				}
			}
			for mac := range ps.MacCountByLoc[loc] {
				ps.Priors[n].MacFreq[loc][mac] = float32(ps.MacCountByLoc[loc][mac]) / float32(maxCount)
				if float64(ps.Priors[n].MacFreq[loc][mac]) < ps.Priors[n].Special["MacFreqMin"] {
					ps.Priors[n].Special["MacFreqMin"] = float64(ps.Priors[n].MacFreq[loc][mac])
				}
			}
		}
	}

	// Deteremine negative mac frequencies and normalize
	for n := range ps.Priors {
		for loc1 := range ps.Priors[n].MacFreq {
			sum := float32(0)
			for loc2 := range ps.Priors[n].MacFreq {
				if loc2 != loc1 {
					for mac := range ps.Priors[n].MacFreq[loc2] {
						ps.Priors[n].NMacFreq[loc1][mac] += ps.Priors[n].MacFreq[loc2][mac]
						sum++
					}
				}
			}
			// Normalize
			if sum > 0 {
				for mac := range ps.Priors[n].MacFreq[loc1] {
					ps.Priors[n].NMacFreq[loc1][mac] = ps.Priors[n].NMacFreq[loc1][mac] / sum
					if float64(ps.Priors[n].NMacFreq[loc1][mac]) < ps.Priors[n].Special["NMacFreqMin"] {
						ps.Priors[n].Special["NMacFreqMin"] = float64(ps.Priors[n].NMacFreq[loc1][mac])
					}
				}
			}
		}
	}

	for n := range ps.Priors {
		ps.Priors[n].Special["MixIn"] = 0.5
		ps.Priors[n].Special["VarabilityCutoff"] = 0
	}

}
