package userbalance

type Transaction struct {
	UserID    int    `json:"userid"`
	Amount    int    `json:"amount"`
	Date      string `json:"date"`
	ServiceID int    `json:"serviceid"`
	OrderID   int    `json:"orderid"`
}

type Money struct {
	FromUserID int    `json:"fromuserid"`
	ToUserID   int    `json:"touserid"`
	Amount     int    `json:"amount"`
	Date       string `json:"date"`
}
