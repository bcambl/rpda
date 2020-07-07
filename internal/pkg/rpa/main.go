package rpa

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

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

	// log.WithFields(log.Fields{
	// 	"method":     method,
	// 	"statusCode": strconv.Itoa(resp.StatusCode),
	// 	"body":       string(body),
	// }).Debug(url)

	return body, resp.StatusCode
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
		if a.Identifiers.ProductionNodeRegexp.MatchString(cs.Name) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	// Non-Production and Non-Test copies in the middle of the slice
	for _, cs := range gcs {
		if a.Identifiers.CopyNodeRegexp.MatchString(cs.Name) &&
			!a.Identifiers.ProductionNodeRegexp.MatchString(cs.Name) &&
			!a.Identifiers.TestNodeRegexp.MatchString(cs.Name) {
			sortedCopiesSettings = append(sortedCopiesSettings, cs)
		}
	}
	// Test copy should be last in slice
	for _, cs := range gcs {
		if a.Identifiers.TestNodeRegexp.MatchString(cs.Name) {
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
	for _, cs := range gcs {
		// always skip the production node
		if a.Identifiers.ProductionNodeRegexp.MatchString(cs.Name) {
			continue
		}
		// return copy settings if an exact copy name was provided matches
		if cs.Name == a.CopyName {
			c = cs
			break
		}
		// there will bo no CopyRegexp set if an exact copy name was provided by user
		if a.CopyRegexp == nil {
			continue
		}
		// check for match against config regex chosen by user
		if a.CopyRegexp.MatchString(cs.Name) {
			// check if the chosen regexp is the test node regex specified by configuration file
			if a.CopyRegexp.String() != a.Identifiers.TestNodeRegexp.String() {
				// was not the test node regex, return if copy does not match the test node regex
				if !a.Identifiers.TestNodeRegexp.MatchString(cs.Name) {
					c = cs
					break
				}
			}
			// return the matching copy
			c = cs
		}
	}
	// when the struct is empty, provide user with valid copies for the consistency group
	if c == (GroupCopiesSettings{}) {
		log.Error("Unable to determine the desired copy to enable direct image access mode")
		if a.CopyName != "" {
			fmt.Println("Requested Copy: ", a.CopyName)
		} else {
			fmt.Println("Requested Copy Regexp: ", a.CopyRegexp.String())
		}
		fmt.Println("Available Copies:")
		for _, cs := range gcs {
			if a.Identifiers.ProductionNodeRegexp.MatchString(cs.Name) {
				// dont print the production node if it matches the production node regexp in config
				continue
			}
			fmt.Println(" - ", cs.Name)
		}
		os.Exit(1)
	}
	return c
}

func (a *App) startTransfer(t Task) error {
	endpoint := fmt.Sprintf(
		a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/clusters/%d/copies/%d/start_transfer",
		t.GroupUID, t.ClusterUID, t.CopyUID)
	if !a.Config.CheckMode {
		body, statusCode := a.apiRequest("PUT", endpoint, nil)
		if statusCode != 204 {
			log.Debugf("Expected status code '204' and received: %d\n", statusCode)
			log.Warnf("Error Starting Transfer for Group %s Copy %s\n", t.GroupName, t.CopyName)
			return errors.New(string(body))
		}
	}
	fmt.Printf("Starting Transfer for Group %s Copy %s\n", t.GroupName, t.CopyName)
	return nil
}

func (a *App) imageAccess(t Task) error {
	operationName := "Disabling"
	operation := "disable_image_access"
	if t.Enable == true {
		operationName = "Enabling"
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

	if !a.Config.CheckMode {
		body, statusCode := a.apiRequest("PUT", endpoint, bytes.NewBuffer(json))
		if statusCode != 204 {
			log.Debugf("Expected status code '204' and received: %d\n", statusCode)
			return errors.New(string(body))
		}
	}
	fmt.Printf("%s Latest Image for Group %s Copy %s\n", operationName, t.GroupName, t.CopyName)
	return nil
}

func (a *App) pollImageAccessEnabled(groupID int, stateDesired bool) {
	pollDelay := 3 // seconds
	pollCount := 0 // iteration counter
	pollMax := 60  // max times to poll before breaking the poll loop
	fmt.Println("waiting for image access to update..")
	groupCopiesSettings := a.getGroupCopiesSettings(groupID)
	copySettings := a.getRequestedCopy(groupCopiesSettings)
	for copySettings.ImageAccessInformation.ImageAccessEnabled != stateDesired {
		log.Debug("current image access enabled: ", copySettings.ImageAccessInformation.ImageAccessEnabled)
		log.Debug("current image logged access mode: ", copySettings.ImageAccessInformation.ImageInformation.Mode)
		time.Sleep(time.Duration(pollDelay) * time.Second)
		groupCopiesSettings = a.getGroupCopiesSettings(groupID)
		copySettings = a.getRequestedCopy(groupCopiesSettings)
		if pollCount > pollMax {
			fmt.Println("Maximum poll count reached while waiting for image access")
			break
		}
		pollCount++
	}
	if stateDesired == true {
		// if the desired state is to have image access == true, we should also
		// ensure that logged access is also set before continuing.
		for copySettings.ImageAccessInformation.ImageInformation.Mode != "LOGGED_ACCESS" {
			log.Debug("current image access enabled: ", copySettings.ImageAccessInformation.ImageAccessEnabled)
			log.Debug("current image logged access mode: ", copySettings.ImageAccessInformation.ImageInformation.Mode)
			time.Sleep(time.Duration(pollDelay) * time.Second)
			groupCopiesSettings = a.getGroupCopiesSettings(groupID)
			copySettings = a.getRequestedCopy(groupCopiesSettings)
			if pollCount > pollMax {
				fmt.Println("Maximum poll count reached while waiting for logged access")
				break
			}
			pollCount++
		}
	}
	log.Debug("current image access enabled: ", copySettings.ImageAccessInformation.ImageAccessEnabled)
	log.Debug("current image logged access mode: ", copySettings.ImageAccessInformation.ImageInformation.Mode)
}

func (a *App) directAccess(t Task) error {
	operationName := "Disabling"
	operation := "disable_direct_access"
	if t.Enable == true {
		operationName = "Enabling"
		operation = "enable_direct_access"
	}
	endpoint := fmt.Sprintf(
		a.Config.RPAURL+"/fapi/rest/5_1/groups/%d/clusters/%d/copies/%d/%s",
		t.GroupUID, t.ClusterUID, t.CopyUID, operation)
	if !a.Config.CheckMode {
		body, statusCode := a.apiRequest("PUT", endpoint, nil)
		if statusCode != 204 {
			log.Debugf("Expected status code '204' and received: %d\n", statusCode)
			log.Warnf("Error enabling Direct Access for Group %s Copy %s\n", t.GroupName, t.CopyName)
			return errors.New(string(body))
		}
	}
	fmt.Printf("%s Direct Access for Group %s Copy %s\n", operationName, t.GroupName, t.CopyName)
	return nil
}

// EnableAll wraper for enabling Direct Image Access for all CG
func (a *App) EnableAll() {
	groups := a.getAllGroups() // only groups user has permission to admin
	for _, g := range groups {
		var t Task
		GroupName := a.getGroupName(g.ID)
		groupCopiesSettings := a.getGroupCopiesSettings(g.ID)
		copySettings := a.getRequestedCopy(groupCopiesSettings)
		// skip if copy is already 'enabled'
		if copySettings.RoleInfo.Role == "ACTIVE" {
			fmt.Printf("Image Access already enabled for %s -> %s\n", a.Group, copySettings.Name)
			continue
		}
		t.GroupName = GroupName
		t.GroupUID = copySettings.CopyUID.GroupUID.ID
		t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
		t.CopyName = copySettings.Name
		t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
		t.Enable = true // whether to enable or disable the following tasks
		if !a.Config.CheckMode {
			err := a.imageAccess(t)
			if err != nil {
				log.Warnf("%s %s\n", GroupName, err)
				continue
			}
			a.pollImageAccessEnabled(g.ID, true)
			err = a.directAccess(t)
			if err != nil {
				log.Warnf("%s %s\n", GroupName, err)
				continue
			}
		}
		time.Sleep(time.Duration(a.Config.Delay) * time.Second)
	}
}

// EnableOne wraper for enabling Direct Image Access for a single CG
func (a *App) EnableOne() {
	groupID := a.getGroupIDByName(a.Group)
	var t Task
	groupCopiesSettings := a.getGroupCopiesSettings(groupID)
	copySettings := a.getRequestedCopy(groupCopiesSettings)
	// skip if copy is already 'enabled'
	if copySettings.RoleInfo.Role == "ACTIVE" {
		fmt.Printf("Image Access already enabled for %s -> %s\n", a.Group, copySettings.Name)
		return
	}
	t.GroupName = a.Group
	t.GroupUID = copySettings.CopyUID.GroupUID.ID
	t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
	t.CopyName = copySettings.Name
	t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
	t.Enable = true // whether to enable or disable the following tasks
	if !a.Config.CheckMode {
		err := a.imageAccess(t)
		if err != nil {
			log.Warnf("%s %s\n", a.Group, err)
			return
		}
		a.pollImageAccessEnabled(groupID, true)
		err = a.directAccess(t)
		if err != nil {
			log.Warnf("%s %s\n", a.Group, err)
		}
	}
}

// FinishAll wraper for finishing Direct Image Access for all CG
func (a *App) FinishAll() {
	groups := a.getAllGroups() // only groups user has permission to admin
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
		t.Enable = false // whether to enable or disable the following tasks
		if !a.Config.CheckMode {
			err := a.imageAccess(t)
			if err != nil {
				log.Warnf("%s %s\n", GroupName, err)
				// continue as we cannot start transfer when image access does
				// not update as expected.
				continue
			}
			a.pollImageAccessEnabled(g.ID, false)
			err = a.startTransfer(t)
			if err != nil {
				log.Warnf("%s %s\n", GroupName, err)
			}
		}
		time.Sleep(time.Duration(a.Config.Delay) * time.Second)
	}
}

// FinishOne wraper for finishing Direct Image Access for a single CG
func (a *App) FinishOne() {
	groupID := a.getGroupIDByName(a.Group)
	var t Task
	groupCopiesSettings := a.getGroupCopiesSettings(groupID)
	copySettings := a.getRequestedCopy(groupCopiesSettings)
	t.GroupName = a.Group
	t.GroupUID = copySettings.CopyUID.GroupUID.ID
	t.ClusterUID = copySettings.CopyUID.GlobalCopyUID.ClusterUID.ID
	t.CopyName = copySettings.Name
	t.CopyUID = copySettings.CopyUID.GlobalCopyUID.CopyUID
	t.Enable = false // whether to enable or disable the following tasks
	if !a.Config.CheckMode {
		err := a.imageAccess(t)
		if err != nil {
			log.Warnf("%s %s\n", a.Group, err)
			return
		}
		a.pollImageAccessEnabled(groupID, false)
		err = a.startTransfer(t)
		if err != nil {
			log.Warnf("%s %s\n", a.Group, err)
		}
	}
}
