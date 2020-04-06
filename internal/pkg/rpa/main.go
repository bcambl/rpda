package rpa

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
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
		fmt.Println("Please Update Configuration File: " + configFile)
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

	examples := `
Package Examples:

// List All Consistency Group Names:
{{ .Bin }} list

// Display status of Test1_CG:
{{ .Bin }} status --group Test1_CG

// Display status of All Consistency Groups:
{{ .Bin }} status --all

// Enable Direct Image Access Mode for Test Copy on Test1_CG
{{ .Bin }} start --group=Test1_CG --latest-test

// Enable Direct Image Access Mode of Test Copy for All Consistency Groups
{{ .Bin }} start --all --latest-test

// Disable Direct Image Access Mode of Test1_CG and Start Transfer
{{ .Bin }} finish --group=Test1_CG

// Disable Direct Image Access Mode of All CG's Start Transfer
{{ .Bin }} finish --all
`
	type usageExampleData struct {
		Bin string
	}
	d := usageExampleData{Bin: os.Args[0]}
	// parse template
	t := template.Must(template.New("usage_examples").Parse(examples))

	// print template to stdout
	err := t.Execute(os.Stdout, &d)
	if err != nil {
		log.Fatal(err)
	}
}

// Debug will dump the
func (a *App) Debug() {

	fmt.Println("DEBUG ENABLED")
	// print out App struct fields
	fmt.Println("RPA URL: ", a.RPAURL)
	fmt.Println("Username: ", a.Username)
	fmt.Println("Password: ", a.Password)
	fmt.Println("Group: ", a.Group)
	fmt.Println("Copy: ", a.Copy)
	fmt.Println("Delay: ", a.Delay)
	fmt.Println("Identifiers:")
	fmt.Println("  Production Node: ", a.Identifiers.ProductionNode)
	fmt.Println("  Copy Node: ", a.Identifiers.CopyNode)
	fmt.Println("  Test Copy: ", a.Identifiers.TestCopy)
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
	endpoint := a.RPAURL + "/fapi/rest/5_1/users/settings/"
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
	endpoint := a.RPAURL + "/fapi/rest/5_1/groups/"
	body, _ := a.apiRequest("GET", endpoint)

	var gResp GroupsResponse
	json.Unmarshal(body, &gResp)
	return gResp.InnerSet
}

func (a *App) getGroupName(groupID int) string {
	endpoint := fmt.Sprintf(a.RPAURL+"/fapi/rest/5_1/groups/%d/name/", groupID)
	body, _ := a.apiRequest("GET", endpoint)

	var groupName GroupName
	json.Unmarshal(body, &groupName)
	return groupName.String
}

func (a *App) getGroupCopiesSettings(groupID int) []GroupCopiesSettings {
	endpoint := fmt.Sprintf(a.RPAURL+"/fapi/rest/5_1/groups/%d/settings/", groupID)
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
		if strings.Contains(cs.Name, a.Identifiers.ProductionNode) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	// Non-Production and Non-Test copies in the middle of the slice
	for _, cs := range gcs {
		if !strings.Contains(cs.Name, a.Identifiers.ProductionNode) &&
			!strings.Contains(cs.Name, a.Identifiers.TestCopy) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	// Test copy should be last in slice
	for _, cs := range gcs {
		if strings.Contains(cs.Name, a.Identifiers.TestCopy) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	return sortedCopiesSettings
}

// ListGroups lists all consistency group names
func (a *App) ListGroups() {
	groups := a.getAllGroups()
	for _, g := range groups {
		name := a.getGroupName(g.ID)
		fmt.Println(name) // consisntency group name
	}
}

// DisplayAllGroups displays the status of all consisntenct groups
func (a *App) DisplayAllGroups() {
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

// DisplayGroup displays the status of a consistency group by group name
func (a *App) DisplayGroup(groupName string) {
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

func (a *App) startTransfer() {
	fmt.Printf("start transfer")
}

func (a *App) imageAccess(enable bool) {
	operation := "disable_image_access"
	if enable {
		operation = "image_access/latest/enable"
	}
	fmt.Printf(operation)
}

func (a *App) directAccess(enable bool) {
	operation := "disable_image_access"
	if enable {
		operation = "enable_direct_access"
	}
	fmt.Printf(operation)
}

// StartAll wraper for enabling Direct Image Access for all CG
func (a *App) StartAll() {
	fmt.Println("enable all to copy: ", a.Copy)
}

// StartOne wraper for enabling Direct Image Access for a single CG
func (a *App) StartOne() {
	fmt.Printf("enable %s to copy: %s\n", a.Group, a.Copy)
}

// FinishAll wraper for finishing Direct Image Access for all CG
func (a *App) FinishAll() {
	fmt.Println("disable all to copy: ", a.Copy)
}

// FinishOne wraper for finishing Direct Image Access for a single CG
func (a *App) FinishOne() {
	fmt.Printf("disable %s to copy: %s\n", a.Group, a.Copy)
}
