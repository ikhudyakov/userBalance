//go:generate easyjson -no_std_marshalers history.go
package models

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
