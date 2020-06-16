package user

import (
	"fmt"
)

// type FrequencyEnum int
// const (
// 	Daily FrequencyEnum = iota
// 	Weekly
// 	Monthly
// 	Yearly
// )

type Cloud struct {
	Name 	string `json:"name"`
	Token 	string `json:"token"`
}

type User struct {
	ID				string		`json:"id"`
	Username		string		`json:"username"`
	Password		string		`json:"password"`
	Frequency		string		`json:"frequency"`
	ToodledoToken	string		`json:"toodledo-token"`
	ToBackup		[]string	`json:"to-backup"`
	Clouds			[]Cloud		`json:"clouds"`
}

func NewUser(name string, pass string, toodledoToken string, toBackup []string) *User {
	u := User{
		ID: "boop",
		Username: name,
		Password: pass,
		Frequency: "daily",
		ToodledoToken: toodledoToken,
		ToBackup: toBackup,
		Clouds: []Cloud{},
	}
	return &u
}

func (u *User) Print() {
	fmt.Println()
	fmt.Println(u.Username)
	fmt.Println(u.ToBackup)
	fmt.Println(u.Frequency)
	fmt.Println()
}
