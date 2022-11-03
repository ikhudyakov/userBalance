package userbalance

type RequestReport struct {
	FromDate string `json:"fromdate"`
	ToDate   string `json:"todate"`
}

type Report struct {
	Title  string
	Amount int
}
