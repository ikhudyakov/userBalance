//go:generate easyjson -no_std_marshalers user.go
package models

import validation "github.com/go-ozzo/ozzo-validation"

//easyjson:json
type User struct {
	Id      int `json:"userid"`
	Balance int `json:"balance"`
}

func (u User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Id,
			validation.Required.Error("id пользователя не может быть не указан либо <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")))
}
