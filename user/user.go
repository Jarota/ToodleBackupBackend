package user

import (
	"fmt"
)

type ToodleInfo struct {
	Token 		string		`json:"token"`
	ToBackup	[]string 	`json:"toBackup"`
}

type Cloud struct {
	Name 	string	`json:"name"`
	Token 	string	`json:"token"`
}

type User struct {
	Username		string
	Password		string
	Frequency		string
	Toodledo 		ToodleInfo
	Clouds			[]Cloud
}

func New(name string, pass string) *User {
	u := User{
		Username: name,
		Password: pass,
		Frequency: "",
		Toodledo: ToodleInfo{},
		Clouds: []Cloud{},
	}
	return &u
}

func (u *User) Print() {
	fmt.Println()
	fmt.Println(u.Username)
	fmt.Println(u.Toodledo)
	fmt.Println(u.Frequency)
	fmt.Println()
}
