package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func slash(c *gin.Context) {
	var group string
	loginGroup := sessions.Default(c)
	groupCookie := loginGroup.Get("group")
	if groupCookie == nil {
		c.Redirect(302, "/login")
	} else {
		group = groupCookie.(string)
		c.Redirect(302, "/dashboard/"+group)
	}
}

func slashLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tmpl", gin.H{})
}

func slashLoginPOST(c *gin.Context) {
	loginGroup := sessions.Default(c)
	group := strings.ToLower(c.PostForm("group"))
	if _, err := os.Stat(path.Join("data", group+".db")); err == nil {
		loginGroup.Set("group", group)
		loginGroup.Save()
		c.Redirect(302, "/dashboard/"+group)
	} else {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"ErrorMessage": "Incorrect login.",
		})
	}
}

func slashLogout(c *gin.Context) {
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
}

func slashDashboard(c *gin.Context) {
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
}

func slashLocation(c *gin.Context) {
	group := c.Param("group")
	user := c.Param("user")
	userJSON := getPositionBreakdown(group, user)
	c.JSON(http.StatusOK, userJSON)
}

func slashExplore(c *gin.Context) {
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
}

func slashPie(c *gin.Context) {
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
}
