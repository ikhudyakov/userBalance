//go:generate easyjson -no_std_marshalers response.go
package models

//easyjson:json
type Response struct {
	Message string `json:"message"`
}
