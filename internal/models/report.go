//go:generate easyjson -no_std_marshalers report.go
package models

//easyjson:json
type (
	RequestReport struct {
		FromDate string `json:"fromdate"`
		ToDate   string `json:"todate"`
	}

	Report struct {
		Title  string
		Amount int
	}
)
