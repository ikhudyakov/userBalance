//go:generate easyjson -no_std_marshalers report.go
package models

import validation "github.com/go-ozzo/ozzo-validation"

//easyjson:json
type (
	RequestReport struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	}

	Report struct {
		Title  string
		Amount int
	}
)

func (r RequestReport) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(
			&r.Month,
			validation.Required.Error("месяца не может быть не указан либо <= 0"),
			validation.Min(1).Error("месяца не может быть <= 0"),
			validation.Max(12).Error("месяца не может быть > 12")),
		validation.Field(
			&r.Year,
			validation.Required.Error("год не может быть не указан либо <= 0"),
			validation.Min(1970).Error("неверно указан год")))
}
