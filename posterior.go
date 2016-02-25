package main

import "math"

func calculatePosterior(res Fingerprint) {
	var ps FullParameters = *NewFullParameters()
	ps, _ = openParameters("find")
	macs := []string{}
	W := make(map[string]int)
	for v2 := range res.WifiFingerprint {
		macs = append(macs, res.WifiFingerprint[v2].Mac)
		W[res.WifiFingerprint[v2].Mac] = res.WifiFingerprint[v2].Rssi
	}
	n, inNetworkAlready := hasNetwork(ps.NetworkMacs, macs)
	Debug.Println(n, inNetworkAlready)
	if !inNetworkAlready {
		Warning.Println("Not in network")
	}

	PBayes1 := make(map[string]float64)
	PBayes2 := make(map[string]float64)
	PA := 1.0 / float64(len(ps.NetworkLocs[n]))
	PnA := (float64(len(ps.NetworkLocs[n])) - 1.0) / float64(len(ps.NetworkLocs[n]))
	for loc := range ps.NetworkLocs[n] {
		PBayes1[loc] = float64(0)
		PBayes2[loc] = float64(0)
		for mac := range W {
			weight := float64(0)
			nweight := float64(0)
			if _, ok := ps.Priors[n].MacFreq[loc][mac]; ok {
				weight = float64(ps.Priors[n].MacFreq[loc][mac])
			} else {
				weight = float64(ps.Priors[n].Special["MacFreqMin"])
				Debug.Println("Using min..")
			}
			if _, ok := ps.Priors[n].NMacFreq[loc][mac]; ok {
				nweight = float64(ps.Priors[n].NMacFreq[loc][mac])
			} else {
				nweight = float64(ps.Priors[n].Special["NMacFreqMin"])
				Debug.Println("Using nmin..")
			}
			PBayes1[loc] += math.Log(weight*PA) - math.Log(weight*PA+PnA*nweight)

			if ps.MacVariability[mac] > 0 {
				ind := int(W[mac] - MinRssi)
				PBA := float64(ps.Priors[n].P[loc][mac][ind])
				PBnA := float64(ps.Priors[n].NP[loc][mac][ind])
				if PBA > 0 {
					PBayes2[loc] += (math.Log(PBA*PA) - math.Log(PBA*PA+PBnA*PnA))
				} else {
					PBayes2[loc] += -1
				}
			}
		}
	}

	Debug.Println(PBayes1)
	Debug.Println(PBayes2)
}
