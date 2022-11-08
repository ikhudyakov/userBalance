package handler

import (
	"net/http"
	"userbalance/internal/service"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "userbalance/docs"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) Init() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", h.getBalance).Methods("POST")
	r.HandleFunc("/topup", h.replenishmentBalance).Methods("POST")
	r.HandleFunc("/transfer", h.transfer).Methods("POST")
	r.HandleFunc("/history", h.getHistory).Methods("POST")
	r.HandleFunc("/report", h.createReport).Methods("POST")
	r.HandleFunc("/reserv", h.reservation).Methods("POST")
	r.HandleFunc("/confirm", h.confirmation).Methods("POST")
	r.HandleFunc("/cancel", h.cancelReservation).Methods("POST")

	fileServer := http.FileServer(http.Dir("./file/"))
	r.PathPrefix("/file/").Handler(http.StripPrefix("/file/", fileServer))

	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	return r
}
