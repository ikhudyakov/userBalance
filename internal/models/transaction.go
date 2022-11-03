//go:generate easyjson -no_std_marshalers transaction.go
package models

//easyjson:json
type (
	Transaction struct {
		UserID    int    `json:"userid"`
		Amount    int    `json:"amount"`
		Date      string `json:"date"`
		ServiceID int    `json:"serviceid"`
		OrderID   int    `json:"orderid"`
	}

	Money struct {
		FromUserID int    `json:"fromuserid"`
		ToUserID   int    `json:"touserid"`
		Amount     int    `json:"amount"`
		Date       string `json:"date"`
	}
)
