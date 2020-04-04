package main

type App struct {
	RPA      string   `json:"rpa_api"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Debug    bool     `json:"-"`
	All      bool     `json:"-"`
	Group    string   `json:"-"`
	Copy     string   `json:"-"`
	Task     string   `json:"-"`
	Delay    int      `json:"delay"`
	CGroups  []string `json:"consistency_groups"`
}

type GroupsResponse struct {
	InnerSet []GroupUID `json:"innerSet"`
}

type GroupUID struct {
	ID int `json:"id"`
}

type GroupName struct {
	String string `json:"string"`
}

type GroupSettingsResponse struct {
	GroupCopiesSettings []GroupCopiesSettings `json:"groupCopiesSettings"`
}

type GroupCopiesSettings struct {
	Name     string   `json:"name"`
	CopyUID  CopyUID  `json:"copyUID"`
	RoleInfo RoleInfo `json:"roleInfo"`
}

type CopyUID struct {
	GroupUID      GroupUID      `json:"groupUID"`
	GlobalCopyUID GlobalCopyUID `json:"globalCopyUID"`
}

type GlobalCopyUID struct {
	CopyUID    int        `json:"copyUID"`
	ClusterUID ClusterUID `json:"clusterUID"`
}

type ClusterUID struct {
	ID int `json:"id"`
}

type RoleInfo struct {
	Role string `json:"role"`
}

type UsersSettingsResponse struct {
	Users []User `json:"users"`
}

type User struct {
	Name   string     `json:"name"`
	Groups []GroupUID `json:"groups"`
}
