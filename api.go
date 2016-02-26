package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

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

		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte("fingerprints-track"))
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				v2 := loadFingerprint(v)
				fmt.Println(string(k), v2.Username)
			}
			return nil
		})

		c.JSON(http.StatusOK, gin.H{"group": jsonData.Group, "user": jsonData.User})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "uhoh"})
	}
}
