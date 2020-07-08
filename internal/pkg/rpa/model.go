package rpa

import "regexp"

// APPLICATION STATE & CONFIGURATION
// =================================================================================================

// App contains settings & variables for the current execution time
type App struct {
	Config      *Config        `json:"config"`
	Group       string         `json:"-"`
	CopyName    string         `json:"-"`
	CopyRegexp  *regexp.Regexp `json:"-"`
	Identifiers *Identifiers   `json:"identifiers"`
}

// Config contains various API configurations for the application
type Config struct {
	RPAURL    string `json:"rpa_url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Delay     int    `json:"delay"`
	PollDelay int    `json:"polldelay"`
	PollMax   int    `json:"pollmax"`
	CheckMode bool   `json:"-"`
	Debug     bool   `json:"-"`
}

// Identifiers describe the regular expression strings for use in copy name validations
type Identifiers struct {
	ProductionNodeRegexp *regexp.Regexp `json:"production_node_regexp"`
	CopyNodeRegexp       *regexp.Regexp `json:"copy_node_regexp"`
	TestNodeRegexp       *regexp.Regexp `json:"test_node_regexp"`
}

// Task is used to pass variables required to perform various tasks to the API
// This helps avoid creating functions with multiple args and provides meaningful variable names
type Task struct {
	GroupName  string
	GroupUID   int
	ClusterUID int
	CopyName   string
	CopyUID    int
	Enable     bool
}

// API RESPONSE DATA STRUCTURES
// =================================================================================================

// UsersSettingsResponse to marshal response from /fapi/rest/5_1/users/settings/
type UsersSettingsResponse struct {
	Users []User `json:"users"`
}

// GroupsResponse to marshal response from /fapi/rest/5_1/groups/
type GroupsResponse struct {
	InnerSet []GroupUID `json:"innerSet"`
}

// GroupSettingsResponse to marshal response from /fapi/rest/5_1/groups/{id}/settings/"
type GroupSettingsResponse struct {
	GroupCopiesSettings []GroupCopiesSettings `json:"groupCopiesSettings"`
}

// User is used by UsersSettingsResponse
type User struct {
	Name   string     `json:"name"`
	Groups []GroupUID `json:"groups"`
}

// GroupUID holds groupUID.id
type GroupUID struct {
	ID int `json:"id"`
}

// GroupName to marshal response from /fapi/rest/5_1/groups/{id}/name/
type GroupName struct {
	String string `json:"string"`
}

// GroupCopiesSettings is used by GroupSettingsResponse for groupCopiesSettings
type GroupCopiesSettings struct {
	Name                   string                 `json:"name"`
	CopyUID                CopyUID                `json:"copyUID"`
	RoleInfo               RoleInfo               `json:"roleInfo"`
	ImageAccessInformation ImageAccessInformation `json:"imageAccessInformation"`
}

// CopyUID is used by GroupCopiesSettings for copyUID within groupCopiesSettings
type CopyUID struct {
	GroupUID      GroupUID      `json:"groupUID"`
	GlobalCopyUID GlobalCopyUID `json:"globalCopyUID"`
}

// GlobalCopyUID is used by GroupCopiesSettings for globalCopyUID within groupCopiesSettings
type GlobalCopyUID struct {
	CopyUID    int        `json:"copyUID"`
	ClusterUID ClusterUID `json:"clusterUID"`
}

// ClusterUID is the same as GroupUID but keeping seperate for clarity and/or if api data expands
type ClusterUID struct {
	ID int `json:"id"`
}

// ImageAccessInformation holds the boolean imageAccessEnabled within groupCopiesSettings
type ImageAccessInformation struct {
	ImageAccessEnabled bool             `json:"imageAccessEnabled"`
	ImageInformation   ImageInformation `json:"imageInformation"`
}

// ImageInformation holds the image information found within ImageAccessInformation
type ImageInformation struct {
	Mode string `json:"mode"`
}

// RoleInfo holds the 'ACTIVE/REPLICA' json string roleInfo within groupCopiesSettings
type RoleInfo struct {
	Role string `json:"role"`
}

// ImageAccessPutData is used to marshal the required PUT data to enable image access
type ImageAccessPutData struct {
	Mode     string `json:"mode"`
	Scenario string `json:"scenario"`
}
