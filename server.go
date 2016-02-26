package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

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

	var ps FullParameters = *NewFullParameters()
	getParameters("findtest2", &ps)
	calculatePriors("findtest2", &ps)
	// fmt.Println(string(dumpParameters(ps)))
	saveParameters("findtest2", ps)
	fmt.Println(ps.MacVariability)
	fmt.Println(ps.NetworkLocs)
	ps, _ = openParameters("findtest2")
	optimizePriors("findtest2")

	fmt.Println(ps.NetworkLocs["0"])

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("static/", "static/")
	r.GET("/dashboard/:group", func(c *gin.Context) {
		group := c.Param("group")
		ps, _ := openParameters(group)
		type DashboardData struct {
			Networks         []string
			Locations        map[string][]string
			LocationAccuracy map[string]int
			LocationCount    map[string]int
		}
		var dash DashboardData
		dash.Networks = []string{}
		dash.Locations = make(map[string][]string)
		dash.LocationAccuracy = make(map[string]int)
		dash.LocationCount = make(map[string]int)
		for n := range ps.NetworkLocs {
			dash.Networks = append(dash.Networks, n)
			dash.Locations[n] = []string{}
			for loc := range ps.NetworkLocs[n] {
				dash.Locations[n] = append(dash.Locations[n], loc)
				dash.LocationAccuracy[loc] = ps.Results[n].Accuracy[loc]
				dash.LocationCount[loc] = ps.Results[n].TotalLocations[loc]
			}
		}
		c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
			"Dash": dash,
		})
	})
	r.GET("/explore/:group/:network/:location", func(c *gin.Context) {
		defer timeTrack(time.Now(), "Loading JSON")
		group := c.Param("group")
		network := c.Param("network")
		location := c.Param("location")
		ps, _ := openParameters(group)
		// TODO: check if network and location exists in the ps, if not return 404
		datas := []template.JS{}
		names := []template.JS{}
		indexNames := []template.JS{}
		it := 0
		for m, n := range ps.Priors[network].P[location] {
			if ps.MacVariability[m] < 0.5 {
				names = append(names, template.JS(string(m)))
				jsonByte, _ := json.Marshal(n)
				datas = append(datas, template.JS(string(jsonByte)))
				indexNames = append(indexNames, template.JS(strconv.Itoa(it)))
				it++
				break
			}
		}
		rsiRange, _ := json.Marshal(RssiRange)
		c.HTML(http.StatusOK, "plot.tmpl", gin.H{
			"RssiRange":  template.JS(string(rsiRange)),
			"Datas":      datas,
			"Names":      names,
			"IndexNames": indexNames,
		})
	})
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
