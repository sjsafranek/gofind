package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/contrib/sessions"
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

// VersionNum keeps track of the version
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

	// var ps FullParameters = *NewFullParameters()
	// getParameters("findtest2", &ps)
	// calculatePriors("findtest2", &ps)
	// // fmt.Println(string(dumpParameters(ps)))
	// saveParameters("findtest2", ps)
	// fmt.Println(ps.MacVariability)
	// fmt.Println(ps.NetworkLocs)
	// optimizePriors("findtest2")
	// ps, _ = openParameters("findtest2")
	//
	// getPositionBreakdown("findtest2", "zack")

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("static/", "static/")
	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/incr", func(c *gin.Context) {
		loginGroup := sessions.Default(c)
		var count int
		v := loginGroup.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count += 1
		}
		loginGroup.Set("count", count)
		loginGroup.Save()
		c.JSON(200, gin.H{"count": count})
	})
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"Message": "You need to login to continue.",
		})
	})
	r.GET("/pie/:group/:network/:location", func(c *gin.Context) {
		group := c.Param("group")
		network := c.Param("network")
		location := c.Param("location")
		ps, _ := openParameters(group)
		vals := []int{}
		names := []string{}
		fmt.Println(ps.Results[network].Guess[location])
		for guessloc, val := range ps.Results[network].Guess[location] {
			names = append(names, guessloc)
			vals = append(vals, val)
		}
		namesJSON, _ := json.Marshal(names)
		valsJSON, _ := json.Marshal(vals)
		c.HTML(http.StatusOK, "pie.tmpl", gin.H{
			"Names": template.JS(namesJSON),
			"Vals":  template.JS(valsJSON),
		})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{})
	})
	r.GET("/", func(c *gin.Context) {
		var group string
		loginGroup := sessions.Default(c)
		groupCookie := loginGroup.Get("group")
		if groupCookie == nil {
			c.Redirect(302, "/login")
		} else {
			group = groupCookie.(string)
			c.Redirect(302, "/dashboard/"+group)
		}
	})
	r.POST("/login", func(c *gin.Context) {
		loginGroup := sessions.Default(c)
		group := c.PostForm("group")
		if _, err := os.Stat(path.Join("data", group+".db")); err == nil {
			loginGroup.Set("group", group)
			loginGroup.Save()
			c.Redirect(302, "/dashboard/"+group)
		} else {
			c.HTML(http.StatusOK, "login.tmpl", gin.H{
				"ErrorMessage": "Incorrect login.",
			})

		}
	})
	r.GET("/logout", func(c *gin.Context) {
		var group string
		loginGroup := sessions.Default(c)
		groupCookie := loginGroup.Get("group")
		if groupCookie == nil {
			c.Redirect(302, "/login")
		} else {
			group = groupCookie.(string)
			fmt.Println(group)
			loginGroup.Clear()
			loginGroup.Save()
			c.HTML(http.StatusOK, "login.tmpl", gin.H{
				"Message": "You are now logged out.",
			})
		}
	})
	r.GET("/dashboard/:group", func(c *gin.Context) {
		group := c.Param("group")
		ps, _ := openParameters(group)
		users := getUsers(group)
		people := make(map[string]UserPositionJSON)
		for _, user := range users {
			people[user] = getPositionBreakdown(group, user)
		}
		type DashboardData struct {
			Networks         []string
			Locations        map[string][]string
			LocationAccuracy map[string]int
			LocationCount    map[string]int
			Mixin            map[string]float64
			VarabilityCutoff map[string]float64
			Users            map[string]UserPositionJSON
		}
		var dash DashboardData
		dash.Networks = []string{}
		dash.Locations = make(map[string][]string)
		dash.LocationAccuracy = make(map[string]int)
		dash.LocationCount = make(map[string]int)
		dash.Mixin = make(map[string]float64)
		dash.VarabilityCutoff = make(map[string]float64)
		for n := range ps.NetworkLocs {
			dash.Mixin[n] = ps.Priors[n].Special["MixIn"]
			dash.VarabilityCutoff[n] = ps.Priors[n].Special["VarabilityCutoff"]
			dash.Networks = append(dash.Networks, n)
			dash.Locations[n] = []string{}
			for loc := range ps.NetworkLocs[n] {
				dash.Locations[n] = append(dash.Locations[n], loc)
				dash.LocationAccuracy[loc] = ps.Results[n].Accuracy[loc]
				dash.LocationCount[loc] = ps.Results[n].TotalLocations[loc]
			}
		}
		c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
			"Group": group,
			"Dash":  dash,
			"Users": people,
		})
	})
	r.GET("/location/:group/:user", func(c *gin.Context) {
		group := c.Param("group")
		user := c.Param("user")
		userJSON := getPositionBreakdown(group, user)
		c.JSON(http.StatusOK, userJSON)
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
		// Sort locations
		macs := []string{}
		for m := range ps.Priors[network].P[location] {
			if float64(ps.MacVariability[m]) > ps.Priors[network].Special["VarabilityCutoff"] {
				macs = append(macs, m)
			}
		}
		sort.Strings(macs)
		it := 0
		for _, m := range macs {
			n := ps.Priors[network].P[location][m]
			names = append(names, template.JS(string(m)))
			jsonByte, _ := json.Marshal(n)
			datas = append(datas, template.JS(string(jsonByte)))
			indexNames = append(indexNames, template.JS(strconv.Itoa(it)))
			it++
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
	r.POST("/track", trackFingerprint)
	r.GET("/whereami", whereAmI)
	r.GET("/editname", editName)
	r.GET("/delete", deleteName)
	r.GET("/calculate", calculate)
	r.GET("/userlocs", userLocations)
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
