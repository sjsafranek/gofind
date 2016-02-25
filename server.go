package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

// RuntimeArgs contains all runtime
// arguments available
var RuntimeArgs struct {
	ExternalIP string
	Port       string
	ServerCRT  string
	ServerKey  string
	SourcePath string
}
var VersionNum string

func main() {
	VersionNum = "0.01"
	// _, executableFile, _, _ := runtime.Caller(0) // get full path of this file
	cwd, _ := os.Getwd()
	flag.StringVar(&RuntimeArgs.Port, "p", ":8003", "port to bind")
	flag.StringVar(&RuntimeArgs.ServerCRT, "crt", "", "location of ssl crt")
	flag.StringVar(&RuntimeArgs.ServerKey, "key", "", "location of ssl key")
	flag.CommandLine.Usage = func() {
		fmt.Println(`find (version ` + VersionNum + `)
run this to start the server and then visit localhost at the port you specify
(see parameters).
Example: 'find yourserver.com'
Example: 'find -p :8080 localhost:8080'
Example: 'find -db /var/lib/find/db.bolt localhost:8003'
Example: 'find -p :8080 -crt ssl/server.crt -key ssl/server.key localhost:8080'
Options:`)
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()
	RuntimeArgs.ExternalIP = flag.Arg(0)
	if RuntimeArgs.ExternalIP == "" {
		RuntimeArgs.ExternalIP = GetLocalIP() + RuntimeArgs.Port
	}
	RuntimeArgs.SourcePath = path.Join(cwd, "data")

	// fmt.Println(getUniqueMacs("find"))
	// fmt.Println(getUniqueLocations("find"))
	makeFullNetwork("test")

	r := gin.Default()
	r.POST("/fingerprint", handleFingerprint)
	r.POST("/learn", handleFingerprint)
	if RuntimeArgs.ServerCRT != "" && RuntimeArgs.ServerKey != "" {
		Info.Println("--------------------------")
		fmt.Println("find (version " + VersionNum + ") is up and running on https://" + RuntimeArgs.ExternalIP)
		fmt.Println("--------------------------")
		r.RunTLS(RuntimeArgs.Port, RuntimeArgs.ServerCRT, RuntimeArgs.ServerKey)
	} else {
		Info.Println("--------------------------")
		Info.Println("find (version " + VersionNum + ") is up and running on http://" + RuntimeArgs.ExternalIP)
		Info.Println("--------------------------")
		r.Run(RuntimeArgs.Port)
	}
}
