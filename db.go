package main

import (
	"log"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

func getUniqueMacs(group string) []string {
	defer timeTrack(time.Now(), "getUniqueMacs")
	uniqueMacs := []string{}

	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			for _, router := range v2.WifiFingerprint {
				if !stringInSlice(router.Mac, uniqueMacs) {
					uniqueMacs = append(uniqueMacs, router.Mac)
				}
			}
		}
		return nil
	})
	return uniqueMacs
}

func getUniqueLocations(group string) []string {
	uniqueLocs := []string{}

	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			if !stringInSlice(v2.Location, uniqueLocs) {
				uniqueLocs = append(uniqueLocs, v2.Location)
			}
		}
		return nil
	})
	return uniqueLocs
}
