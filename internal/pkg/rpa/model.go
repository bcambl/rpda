package rpa

// APPLICATION STATE & CONFIGURATION
// =================================================================================================

// App contains settings & variables for the current execution time
type App struct {
	RPAURL      string      `json:"rpa_url"`
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	Group       string      `json:"-"`
	Copy        string      `json:"-"`
	Delay       int         `json:"delay"`
	Identifiers Identifiers `json:"identifiers"`
}

// Identifiers describe the prefix/suffix strings for use in a 'contains' query
type Identifiers struct {
	ProductionNode string `json:"production_node_name_contains"`
	CopyNode       string `json:"dr_copy_name_contains"`
	TestCopy       string `json:"test_copy_name_contains"`
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

// GroupName to marchal response from /fapi/rest/5_1/groups/{id}/name/
type GroupName struct {
	String string `json:"string"`
}

// GroupCopiesSettings is used by GroupSettingsResponse for groupCopiesSettings
type GroupCopiesSettings struct {
	Name     string   `json:"name"`
	CopyUID  CopyUID  `json:"copyUID"`
	RoleInfo RoleInfo `json:"roleInfo"`
}

// CopyUID is used by GroupCopiesSettings for copyIUD within groupCopiesSettings
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

// RoleInfo holds the 'ACTIVE/REPLICA' json string roleInfo within groupCopiesSettings
type RoleInfo struct {
	Role string `json:"role"`
}
