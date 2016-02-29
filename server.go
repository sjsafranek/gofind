package gofind

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

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
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"ErrorMessage": "Please login first.",
		})
	})

	// webpages (routes.go)
	r.GET("/pie/:group/:network/:location", slashPie)
	r.GET("/", slash)
	r.GET("/login", slashLogin)
	r.POST("/login", slashLoginPOST)
	r.GET("/logout", slashLogout)
	r.GET("/dashboard/:group", slashDashboard)
	r.GET("/location/:group/:user", slashLocation)
	r.GET("/explore/:group/:network/:location", slashExplore)

	// fingerprinting stuff (fingerprint.go)
	r.POST("/fingerprint", handleFingerprint)
	r.POST("/learn", handleFingerprint)
	r.POST("/track", trackFingerprint)

	// API routes (api.go)
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
