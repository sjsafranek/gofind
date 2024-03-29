package main

import (
	"fmt"
	"path"
	"testing"

	"github.com/boltdb/bolt"
)

func BenchmarkLoadParameters(b *testing.B) {
	var ps FullParameters = *NewFullParameters()
	db, err := bolt.Open(path.Join("data", "testdb.db"), 0600, nil)
	if err != nil {
		Error.Println(err)
	}
	defer db.Close()
	for i := 0; i < b.N; i++ {
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
		if err != nil {
			Error.Println(err)
		}

	}
}
