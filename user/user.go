package user

import (
	"fmt"
)

type Cloud struct {
	Name 	string	`json:"name"`
	Token 	string	`json:"token"`
}

type User struct {
	Username		string
	Password		string
	Frequency		string
	ToodledoToken	string
	ToBackup		[]string
	Clouds			[]Cloud
}

func New(name string, pass string) *User {
	u := User{
		Username: name,
		Password: pass,
		Frequency: "",
		ToodledoToken: "",
		ToBackup: []string{},
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
