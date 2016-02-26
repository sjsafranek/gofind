package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

func calculate(c *gin.Context) {
	group := c.DefaultQuery("group", "noneasdf")
	if group != "noneasdf" {
		optimizePriors(group)
		c.JSON(http.StatusOK, gin.H{"message": "Parameters optimized.", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func editName(c *gin.Context) {
	group := c.DefaultQuery("group", "noneasdf")
	location := c.DefaultQuery("location", "none")
	newname := c.DefaultQuery("newname", "none")
	if group != "noneasdf" {
		toUpdate := make(map[string]string)
		numChanges := 0

		db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
		if err != nil {
			log.Fatal(err)
		}

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("fingerprints"))
			if b != nil {
				c := b.Cursor()
				for k, v := c.Last(); k != nil; k, v = c.Prev() {
					v2 := loadFingerprint(v)
					if v2.Location == location {
						v2.Location = newname
						toUpdate[string(k)] = string(dumpFingerprint(v2))
					}
				}
			}
			return nil
		})

		db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			for k, v := range toUpdate {
				bucket.Put([]byte(k), []byte(v))
			}
			return nil
		})

		numChanges += len(toUpdate)

		toUpdate = make(map[string]string)

		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("fingerprints-track"))
			if b != nil {
				c := b.Cursor()
				for k, v := c.Last(); k != nil; k, v = c.Prev() {
					v2 := loadFingerprint(v)
					if v2.Location == location {
						v2.Location = newname
						toUpdate[string(k)] = string(dumpFingerprint(v2))
					}
				}
			}
			return nil
		})

		db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("fingerprints-track"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			for k, v := range toUpdate {
				bucket.Put([]byte(k), []byte(v))
			}
			return nil
		})

		db.Close()
		numChanges += len(toUpdate)
		regenerateEverything(group)

		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

func deleteName(c *gin.Context) {
	group := c.DefaultQuery("group", "noneasdf")
	location := c.DefaultQuery("location", "none")
	if group != "noneasdf" {
		numChanges := 0

		db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, group+".db"), 0600, nil)
		if err != nil {
			log.Fatal(err)
		}

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("fingerprints"))
			if b != nil {
				c := b.Cursor()
				for k, v := c.Last(); k != nil; k, v = c.Prev() {
					v2 := loadFingerprint(v)
					if v2.Location == location {
						b.Delete(k)
						numChanges++
					}
				}
			}
			return nil
		})

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("fingerprints"))
			if b != nil {
				c := b.Cursor()
				for k, v := c.Last(); k != nil; k, v = c.Prev() {
					v2 := loadFingerprint(v)
					if v2.Location == location {
						b.Delete(k)
						numChanges++
					}
				}
			}
			return nil
		})

		db.Close()
		regenerateEverything(group)

		c.JSON(http.StatusOK, gin.H{"message": "Changed name of " + strconv.Itoa(numChanges) + " things", "success": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Error parsing request"})
	}
}

type WhereAmIJson struct {
	Group string `json:"group"`
	User  string `json:"user"`
}

func whereAmI(c *gin.Context) {
	var jsonData WhereAmIJson
	if c.BindJSON(&jsonData) == nil {
		defer timeTrack(time.Now(), "getUniqueMacs")
		db, err := bolt.Open(path.Join(RuntimeArgs.SourcePath, jsonData.Group+".db"), 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		locations := []string{}
		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte("fingerprints-track"))
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := loadFingerprint(v)
				fmt.Println(string(k), v2.Username)
				if v2.Username == jsonData.User {
					locations = append(locations, v2.Location)
				}
				if len(locations) > 2 {
					break
				}
			}
			return nil
		})
		// jsonLocations, _ := json.Marshal(locations)
		c.JSON(http.StatusOK, gin.H{"group": jsonData.Group, "user": jsonData.User, "locations": locations})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "uhoh"})
	}
}
