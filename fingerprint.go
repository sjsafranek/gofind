package main

import (
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"path"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

// Fingerprint is the prototypical information from the fingerprinting device
type Fingerprint struct {
	Group           string   `json:"group"`
	Username        string   `json:"username"`
	Location        string   `json:"location"`
	WifiFingerprint []Router `json:"wifi-fingerprint"`
}

// Router is the router information for each invdividual mac address
type Router struct {
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

func dumpFingerprint(res Fingerprint) []byte {
	dumped, _ := json.Marshal(res)
	return dumped
}

func loadFingerprint(jsonByte []byte) Fingerprint {
	res := Fingerprint{}
	json.Unmarshal(jsonByte, &res)
	return res
}

func cleanFingerprint(res *Fingerprint) {
	res.Group = strings.ToLower(res.Group)
	res.Location = strings.ToLower(res.Location)
	res.Username = strings.ToLower(res.Username)
	for r := range res.WifiFingerprint {
		if res.WifiFingerprint[r].Rssi >= 0 {
			res.WifiFingerprint[r].Rssi = int(res.WifiFingerprint[r].Rssi/2) - 100
		}
	}
}

func putFingerprintIntoDatabase(res Fingerprint, database string) error {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(database))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		timestamp := time.Now().Format(time.StampMilli)
		err = bucket.Put([]byte(timestamp), dumpFingerprint(res))
		if err != nil {
			return fmt.Errorf("could add to bucket: %s", err)
		}
		return err
	})
	return err
}

func handleFingerprint(c *gin.Context) {
	var jsonFingerprint Fingerprint
	if c.BindJSON(&jsonFingerprint) == nil {
		cleanFingerprint(&jsonFingerprint)
		if jsonFingerprint.Location != "" {
			putFingerprintIntoDatabase(jsonFingerprint, "fingerprints")
			Debug.Println("Inserted fingerprint for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location)
			c.JSON(http.StatusOK, gin.H{"message": "Inserted fingerprint"})
		} else {
			trackFingerprint(c)
		}
	}
}

func trackFingerprint(c *gin.Context) {
	var jsonFingerprint Fingerprint
	if c.BindJSON(&jsonFingerprint) == nil {
		cleanFingerprint(&jsonFingerprint)
		locationGuess, _ := calculatePosterior(jsonFingerprint, *NewFullParameters())
		jsonFingerprint.Location = locationGuess
		putFingerprintIntoDatabase(jsonFingerprint, "fingerprints-track")
		positions := [][]string{}
		positions1 := []string{}
		positions2 := []string{}
		positions1 = append(positions1, locationGuess)
		positions2 = append(positions2, " ")
		positions = append(positions, positions1)
		positions = append(positions, positions2)
		Debug.Println("Tracking fingerprint for " + jsonFingerprint.Username + " (" + jsonFingerprint.Group + ") at " + jsonFingerprint.Location + " (guess)")
		c.JSON(http.StatusOK, gin.H{"message": "Inserted fingerprint", "locationGuess": locationGuess, "position": positions})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "UH OH"})
	}
}
