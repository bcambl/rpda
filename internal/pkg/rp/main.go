package rp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Load will populate the application config via vobra/viper
func (c *Config) Load(cmd *cobra.Command) *Config {
	c.RPAURL = viper.GetString("api.url")
	c.Username = viper.GetString("api.username")
	c.Password = viper.GetString("api.password")
	c.Delay = viper.GetInt("api.delay")
	c.NoOp = viper.GetBool("noop")
	c.Debug = viper.GetBool("debug")
	return c
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

// Debugger will dump the
func (a *App) Debugger() {

	fmt.Println("DEBUG ENABLED")
	// print out App struct fields
	fmt.Println("RPA URL: ", a.Config.RPAURL)
	fmt.Println("Username: ", a.Config.Username)
	fmt.Println("Password: ", a.Config.Password)
	fmt.Println("Group: ", a.Group)
	fmt.Println("Copy: ", a.Copy)
	fmt.Println("Delay: ", a.Config.Delay)
	fmt.Println("Debug: ", a.Config.Debug)
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

func (a *App) apiRequest(method, url string, data io.Reader) ([]byte, int) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout: 1 * time.Second,
	}
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		log.Fatal(err)
	}
	authString := basicAuth(a.Config.Username, a.Config.Password)
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
	endpoint := a.Config.RPAURL + "/fapi/rest/5_1/users/settings/"
	body, _ := a.apiRequest("GET", endpoint, nil)
	var usr UsersSettingsResponse
	json.Unmarshal(body, &usr)

	var allowedGroups []GroupUID
	for _, u := range usr.Users {
		if u.Name == a.Config.Username {
			allowedGroups = u.Groups
		}
	}
	return allowedGroups
}

func (a *App) userHasGrouAdmin(groupID int, usersGroups []GroupUID) bool {
	var permission bool
	if usersGroups == nil {
		usersGroups = a.getUserGroups()
	}
	for _, g := range usersGroups {
		if g.ID == groupID {
			permission = true
		}
	}
	return permission
}

func (a *App) getAllGroups() []GroupUID {
	endpoint := a.Config.RPAURL + "/fapi/rest/5_1/groups/"
	body, _ := a.apiRequest("GET", endpoint, nil)

	var gResp GroupsResponse
	json.Unmarshal(body, &gResp)
	return gResp.InnerSet
}

func (a *App) getGroupName(groupID int) string {
	endpoint := fmt.Sprintf(a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/name/", groupID)
	body, _ := a.apiRequest("GET", endpoint, nil)

	var groupName GroupName
	json.Unmarshal(body, &groupName)
	return groupName.String
}

func (a *App) getGroupIDByName(groupName string) int {
	var id int
	allGroups := a.getAllGroups()
	for _, g := range allGroups {
		n := a.getGroupName(g.ID)
		if groupName == n {
			id = g.ID
		}
	}
	return id
}

func (a *App) getGroupCopiesSettings(groupID int) []GroupCopiesSettings {
	endpoint := fmt.Sprintf(a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/settings/", groupID)
	body, _ := a.apiRequest("GET", endpoint, nil)

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
	groupID := a.getGroupIDByName(groupName)
	fmt.Println(groupName) // consisntency group name
	copySettings := a.getGroupCopiesSettings(groupID)
	for _, cs := range copySettings {
		fmt.Printf("\t%s (%s)\n", cs.Name, cs.RoleInfo.Role)
	}
}

// getRequestedCopy attempts to determine the desired copy based on identifier prefixes and flags
func (a *App) getRequestedCopy(gcs []GroupCopiesSettings) GroupCopiesSettings {
	var c GroupCopiesSettings
	for _, cs := range gcs { // iterate over all copies
		if strings.Contains(cs.Name, a.Copy) && // copy contains desired criteria
			!strings.Contains(cs.Name, a.Identifiers.ProductionNode) { // & NOT the production node
			if a.Copy != a.Identifiers.TestCopy { // if desired copy does not match test identifier
				if !strings.Contains(cs.Name, a.Identifiers.TestCopy) { // skip the test identifier
					c = cs
				}
			} else {
				c = cs
			}
		}
	}
	if c == (GroupCopiesSettings{}) {
		log.Fatal("Unable to determine the desired copy to enable direct image access mode")
	}
	return c
}

func (a *App) startTransfer(t Task) {
	endpoint := fmt.Sprintf(
		a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/clusters/%d/copies/%d/start_transfer",
		t.GroupUID, t.ClusterUID, t.CopyUID)
	_, statusCode := a.apiRequest("PUT", endpoint, nil)
	if statusCode != 204 {
		log.Fatalf("Error STARTING TRANSFER for Group %s Copy %s\n", t.GroupName, t.CopyName)
	}
}

func (a *App) imageAccess(t Task) {
	operation := "disable_image_access"
	if t.Enable == true {
		operation = "image_access/latest/enable"
	}
	endpoint := fmt.Sprintf(
		a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/clusters/%d/copies/%d/%s",
		t.GroupUID, t.ClusterUID, t.CopyUID, operation)

	var d ImageAccessPutData
	d.Mode = "LOGGED_ACCESS"
	d.Scenario = "UNKNOWN"

	json, err := json.Marshal(&d)
	if err != nil {
		log.Fatal(err)
	}

	_, statusCode := a.apiRequest("PUT", endpoint, bytes.NewBuffer(json))
	if statusCode != 204 {
		log.Fatalf("Error enabling LATEST IMAGE for Group %s Copy %s\n", t.GroupName, t.CopyName)
	}
}

func (a *App) directAccess(t Task) {
	operation := "disable_direct_access"
	if t.Enable == true {
		operation = "enable_direct_access"
	}
	endpoint := fmt.Sprintf(
		a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/clusters/%d/copies/%d/%s",
		t.GroupUID, t.ClusterUID, t.CopyUID, operation)
	_, statusCode := a.apiRequest("PUT", endpoint, nil)
	if statusCode != 204 {
		log.Fatalf("Error enabling DIRECT ACCESS for Group %s Copy %s\n", t.GroupName, t.CopyName)
	}
}

// StartAll wraper for enabling Direct Image Access for all CG
func (a *App) StartAll() {
	groups := a.getUserGroups() // only groups user has permission to admin
	for _, g := range groups {
		var t Task
		GroupName := a.getGroupName(g.ID)
		groupCopiesSettings := a.getGroupCopiesSettings(g.ID)
		copySettings := a.getRequestedCopy(groupCopiesSettings)
		t.GroupName = GroupName
		t.GroupUID = copySettings.CopyUID.GroupUID.ID
		t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
		t.CopyName = copySettings.Name
		t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
		t.Enable = true // whether to enable or disable the following tasks
		a.imageAccess(t)
		time.Sleep(3 * time.Second) // wait a few seconds for platform
		a.directAccess(t)
		fmt.Printf("enabled direct image access for %s on copy %s\n", GroupName, copySettings.Name)
	}
}

// StartOne wraper for enabling Direct Image Access for a single CG
func (a *App) StartOne() {
	groupID := a.getGroupIDByName(a.Group)
	usersGroups := a.getUserGroups()
	if a.userHasGrouAdmin(groupID, usersGroups) == false {
		log.Error("User does not have sufficient access to administer ", a.Group)
		return
	}
	var t Task
	groupCopiesSettings := a.getGroupCopiesSettings(groupID)
	copySettings := a.getRequestedCopy(groupCopiesSettings)
	t.GroupName = a.Group
	t.GroupUID = copySettings.CopyUID.GroupUID.ID
	t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
	t.CopyName = copySettings.Name
	t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
	t.Enable = true // whether to enable or disable the following tasks
	a.imageAccess(t)
	time.Sleep(3 * time.Second) // wait a few seconds for platform
	a.directAccess(t)
	fmt.Printf("enabled direct image access for %s on copy %s\n", a.Group, copySettings.Name)
}

// FinishAll wraper for finishing Direct Image Access for all CG
func (a *App) FinishAll() {
	groups := a.getUserGroups() // only groups user has permission to admin
	for _, g := range groups {
		var t Task
		GroupName := a.getGroupName(g.ID)
		groupID := a.getGroupIDByName(a.Group)
		groupCopiesSettings := a.getGroupCopiesSettings(groupID)
		copySettings := a.getRequestedCopy(groupCopiesSettings)
		t.GroupName = GroupName
		t.GroupUID = copySettings.CopyUID.GroupUID.ID
		t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
		t.CopyName = copySettings.Name
		t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
		t.Enable = false // whether to enable or disable the following tasks
		a.imageAccess(t)
		time.Sleep(3 * time.Second) // wait a few seconds for platform
		a.startTransfer(t)
		fmt.Printf("finished direct image access for %s on copy %s\n", GroupName, copySettings.Name)
	}
}

// FinishOne wraper for finishing Direct Image Access for a single CG
func (a *App) FinishOne() {
	groupID := a.getGroupIDByName(a.Group)
	usersGroups := a.getUserGroups()
	if a.userHasGrouAdmin(groupID, usersGroups) == false {
		log.Error("User does not have sufficient access to administer ", a.Group)
		return
	}
	var t Task
	groupCopiesSettings := a.getGroupCopiesSettings(groupID)
	copySettings := a.getRequestedCopy(groupCopiesSettings)
	t.GroupName = a.Group
	t.GroupUID = copySettings.CopyUID.GroupUID.ID
	t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
	t.CopyName = copySettings.Name
	t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
	t.Enable = false // whether to enable or disable the following tasks
	a.imageAccess(t)
	time.Sleep(3 * time.Second) // wait a few seconds for platform
	a.startTransfer(t)
	fmt.Printf("finished direct image access for %s on copy %s\n", a.Group, copySettings.Name)
}
