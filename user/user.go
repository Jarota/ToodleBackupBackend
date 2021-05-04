package user

import (
	"fmt"
	"time"
)

// ToodleInfo type to contain toodledo token and permissions
type ToodleInfo struct {
	Token    string   `json:"token"`
	Refresh  string   `json:"refresh"`
	ToBackup []string `json:"toBackup"`
}

// Cloud type to contain cloud service token
type Cloud struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// BackupTime describes the time at which the user's data should be backed up
type BackupTime struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`
}

// User type containing all user info - Time is the time to backup the data
type User struct {
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Frequency string     `json:"frequency"`
	Time      BackupTime `json:"time"`
	Toodledo  ToodleInfo `json:"toodledo"`
	Clouds    []Cloud    `json:"clouds"`
}

// New creates a new skeleton user from a username and password
func New(name string, pass string) *User {
	h, m, s := time.Now().UTC().Clock()
	now := BackupTime{Hour: h, Minute: m, Second: s}
	u := User{
		Username:  name,
		Password:  pass,
		Frequency: "",
		Time:      now,
		Toodledo:  ToodleInfo{Token: "", Refresh: "", ToBackup: []string{}},
		Clouds:    []Cloud{},
	}
	return &u
}

// Print - certain attributes of a given user
func (u *User) Print() {
	fmt.Println()
	fmt.Println(u.Username)
	fmt.Println(u.Toodledo)
	fmt.Println(u.Frequency)
	fmt.Println()
}
