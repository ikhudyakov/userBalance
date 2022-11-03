//go:generate easyjson -no_std_marshalers user.go
package models

//easyjson:json
type User struct {
	Id      int `json:"userid"`
	Balance int `json:"balance"`
}
