package main

import (
	"log"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

func getUsers(group string) []string {
	defer timeTrack(time.Now(), "getUsers")
	if _, ok := usersCache[group]; ok {
		return usersCache[group]
	}

	uniqueUsers := []string{}

	db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			if !stringInSlice(v2.Username, uniqueUsers) {
				uniqueUsers = append(uniqueUsers, v2.Username)
			}
		}
		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("fingerprints-track"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			v2 := loadFingerprint(v)
			if !stringInSlice(v2.Username, uniqueUsers) {
				uniqueUsers = append(uniqueUsers, v2.Username)
			}
		}
		return nil
	})
	usersCache[group] = uniqueUsers
	return uniqueUsers
}

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

func getUniqueLocations(group string) (uniqueLocs []string) {
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

func getMacCount(group string) (macCount map[string]int) {
	macCount = make(map[string]int)
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
				if _, ok := macCount[router.Mac]; !ok {
					macCount[router.Mac] = 0
				}
				macCount[router.Mac]++
			}
		}
		return nil
	})
	return
}

func getMacCountByLoc(group string) (macCountByLoc map[string]map[string]int) {
	macCountByLoc = make(map[string]map[string]int)
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
			if _, ok := macCountByLoc[v2.Location]; !ok {
				macCountByLoc[v2.Location] = make(map[string]int)
			}
			for _, router := range v2.WifiFingerprint {
				if _, ok := macCountByLoc[v2.Location][router.Mac]; !ok {
					macCountByLoc[v2.Location][router.Mac] = 0
				}
				macCountByLoc[v2.Location][router.Mac]++
			}
		}
		return nil
	})
	return
}
