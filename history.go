package userbalance

type History struct {
	Date        string
	Amount      int
	Description string
}

type RequestHistory struct {
	UserID    int    `json:"userid"`
	SortField string `json:"sortfield"`
	Direction string `json:"direction"`
}
