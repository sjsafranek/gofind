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

type Fingerprint struct {
	Group           string   `json:"group"`
	Username        string   `json:"username"`
	Location        string   `json:"location"`
	WifiFingerprint []Router `json:"wifi-fingerprint"`
}

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
}

func putFingerprintIntoDatabase(res Fingerprint) error {
	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, res.Group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
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

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
	return err

}

func handleFingerprint(c *gin.Context) {
	var json Fingerprint
	if c.BindJSON(&json) == nil {
		cleanFingerprint(&json)
		if json.Location != "" {
			putFingerprintIntoDatabase(json)
			c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "your current location is XYZ"})
		}
	}
}
