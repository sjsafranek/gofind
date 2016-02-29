package main

// func ExampleF_testJSON() {
// 	db, err := bolt.Open("data/find.db", 0600, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()
// 	iLimit := 3
// 	i := 0
// 	db.View(func(tx *bolt.Tx) error {
// 		// Assume bucket exists and has keys
// 		b := tx.Bucket([]byte("fingerprints"))
// 		c := b.Cursor()
// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			i++
// 			if i > iLimit {
// 				fingerprintJSON := loadFingerprint(v)
// 				fmt.Println(string(dumpFingerprint(fingerprintJSON)))
// 				// calculatePosterior(fingerprintJSON)
// 				break
// 			}
// 		}
// 		return nil
// 	})
// 	// Output: TLKYnv64wOquExOTAyMjAxNg
// }
//
// func init() {
// 	RuntimeArgs.SourcePath = "data"
// }
