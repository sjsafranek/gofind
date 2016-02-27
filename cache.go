package main

import "time"

var psCache map[string]FullParameters
var usersCache map[string][]string

func init() {
	go clearCache()
}

func clearCache() {
	for {
		Debug.Println("Clearing cache.")
		psCache = make(map[string]FullParameters)
		usersCache = make(map[string][]string)
		time.Sleep(time.Minute * 10)
	}
}
