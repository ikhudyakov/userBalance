//go:generate easyjson -no_std_marshalers history.go
package models

import validation "github.com/go-ozzo/ozzo-validation"

//easyjson:json
type (
	History struct {
		Date        string `json:"date"`
		Amount      int    `json:"amount"`
		Description string `json:"description"`
	}

	Histories struct {
		Entity []History `json:"entity"`
	}

	RequestHistory struct {
		UserID    int    `json:"userid"`
		SortField string `json:"sortfield"`
		Direction string `json:"direction"`
	}
)

func (r RequestHistory) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(
			&r.UserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")))
}
