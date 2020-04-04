package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	// ProductionIdentifier is the Production Node Identifier
	ProductionIdentifier = "_PN"
	// CopyIdentifier is the Copy Copy Identifier
	CopyIdentifier = "_CN"
	// TestIdentifier is the Test Copy Identifier
	TestIdentifier = "TC_"
)

func (a *App) readConfig(configFile string) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Configuration File not found.. creating\n")
		byteArray, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(configFile, byteArray, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Please Update Configuration File: " + configFile)
		os.Exit(0)
	}
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(config, &a)
}

func (a *App) usageExamples() {
	flag.Usage()
	msg := `
CLI Examples:

// Display status of Test1_CG:
program -task=display -group=Test1_CG

// Display status of All Consistency Groups:
program -task=display -all

// Enable Direct Image Access Mode for Test Copy on Test1_CG
program -task=enable -copy=TEST -group=Test1_CG 

// Enable Direct Image Access Mode of Test Copy for All Consistency Groups
program -task=enable -copy=TEST -all

// Disable Direct Image Access Mode of Test1_CG and Start Transfer
program -task=disable -group=Test1_CG

// Disable Direct Image Access Mode of All CG's Start Transfer
program -task=disable -all
`
	fmt.Println(msg)
}

func (a *App) debug() {
	// // Show all Group ID's
	// groups := GetAllGroups()
	// for _, g := range groups {
	// 	name := GetGroupName(g.ID)
	// 	fmt.Printf("%s: %d\n", name, g.ID)
	// }

	// //Show Test1_CG
	// copySettings := GetGroupCopiesSettings(1157507498)
	// for _, cs := range copySettings {
	// 	fmt.Println("Consistency Group Name: ", cs.Name)
	// 	fmt.Println("GroupUID: ", cs.CopyUID.GroupUID.ID)
	// 	fmt.Println("ClusterID: ", cs.CopyUID.GlobalCopyUID.ClusterUID.ID)
	// 	fmt.Println("GlobalCopy CopyUID: ", cs.CopyUID.GlobalCopyUID.CopyUID)
	// 	fmt.Println("Role: ", cs.RoleInfo.Role)
	// 	fmt.Println("===")
	// }

	// DisplayAllGroups()
	// DisplayGroup("Test1_CG")
	fmt.Println("DEBUG ENABLED")
	// debug out cli args
	fmt.Println("RPA URL: ", a.RPA)
	fmt.Println("Username: ", a.Username)
	fmt.Println("Task: ", a.Task)
	fmt.Println("Group: ", a.Group)
	fmt.Println("Copy: ", a.Copy)
	fmt.Println("All: ", a.All)
	fmt.Println("Delay: ", a.Delay)
	fmt.Println("Debug: ", a.Debug)
	fmt.Println("Uncaught Arguments: ", flag.Args())
}

func basicAuth(username, password string) string {
	userPass := username + ":" + password
	b64String := base64.StdEncoding.EncodeToString([]byte(userPass))
	authString := "Basic " + b64String
	return authString
}

func (a *App) apiRequest(method, url string) ([]byte, int) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout: 1 * time.Second,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	authString := basicAuth(a.Username, a.Password)
	req.Header.Set("Authorization", authString)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	return body, resp.StatusCode
}

func (a *App) getUserGroups() []GroupUID {
	endpoint := a.RPA + "/fapi/rest/5_1/users/settings/"
	body, statusCode := a.apiRequest("GET", endpoint)
	fmt.Println(statusCode)
	var usr UsersSettingsResponse
	json.Unmarshal(body, &usr)

	var allowedGroups []GroupUID
	for _, u := range usr.Users {
		if u.Name == a.Username {
			allowedGroups = u.Groups
		}
	}
	return allowedGroups
}

func (a *App) getAllGroups() []GroupUID {
	endpoint := a.RPA + "/fapi/rest/5_1/groups/"
	body, _ := a.apiRequest("GET", endpoint)

	var gResp GroupsResponse
	json.Unmarshal(body, &gResp)
	return gResp.InnerSet
}

func (a *App) getGroupName(groupID int) string {
	endpoint := fmt.Sprintf(a.RPA+"/fapi/rest/5_1/groups/%d/name/", groupID)
	body, _ := a.apiRequest("GET", endpoint)

	var groupName GroupName
	json.Unmarshal(body, &groupName)
	return groupName.String
}

func (a *App) getGroupCopiesSettings(groupID int) []GroupCopiesSettings {
	endpoint := fmt.Sprintf(a.RPA+"/fapi/rest/5_1/groups/%d/settings/", groupID)
	body, _ := a.apiRequest("GET", endpoint)

	var gsr GroupSettingsResponse
	json.Unmarshal(body, &gsr)
	result := a.sortGroupCopies(gsr.GroupCopiesSettings)
	return result
}

func (a *App) sortGroupCopies(gcs []GroupCopiesSettings) []GroupCopiesSettings {
	var sortedCopiesSettings []GroupCopiesSettings
	// Production should be index 0
	for _, cs := range gcs {
		if strings.Contains(cs.Name, ProductionIdentifier) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	// Non-Production and Non-Test copies in the middle of the slice
	for _, cs := range gcs {
		if !strings.Contains(cs.Name, ProductionIdentifier) &&
			!strings.Contains(cs.Name, TestIdentifier) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	// Test copy should be last in slice
	for _, cs := range gcs {
		if strings.Contains(cs.Name, TestIdentifier) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	return sortedCopiesSettings
}

func (a *App) displayAllGroups() {
	groups := a.getAllGroups()
	for _, g := range groups {
		name := a.getGroupName(g.ID)
		fmt.Println(name) // consisntency group name
		copySettings := a.getGroupCopiesSettings(g.ID)
		for _, cs := range copySettings {
			fmt.Printf("\t%s (%s)\n", cs.Name, cs.RoleInfo.Role)
		}
	}
}

func (a *App) displayGroup(groupName string) {
	groups := a.getAllGroups()
	for _, g := range groups {
		name := a.getGroupName(g.ID)
		if groupName == name {
			fmt.Println(name) // consisntency group name
			copySettings := a.getGroupCopiesSettings(g.ID)
			for _, cs := range copySettings {
				fmt.Printf("\t%s (%s)\n", cs.Name, cs.RoleInfo.Role)
			}
			break
		}
	}
}

// Run is the Main application routine
func (a *App) Run() {
	if a.All {
		fmt.Println("Operating on: All Consistency Groups")
	} else {
		fmt.Println("Operating on Group: ", a.Group)
	}

	if a.Task == "display" {
		if a.All {
			a.displayAllGroups()
			os.Exit(0)
		}
		a.displayGroup(a.Group)
		os.Exit(0)
	}
}
