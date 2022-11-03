package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"userbalance"
)

// @Summary Get Balance
// @Tags balance
// @Description getting the user's balance
// @ID get-balance
// @Accept  json
// @Produce  json
// @Param input body userbalance.User true "user id"
// @Success 200 {object} userbalance.User
// @Failure 500 {object} userbalance.Response
// @Router / [post]
func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var userId int
	var err error
	var user userbalance.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	userId = user.Id

	if user, err = h.services.GetBalance(userId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// @Summary Replenishment Balance
// @Tags balance
// @Description replenishment of the user's balance
// @ID replenishment-balance
// @Accept  json
// @Produce  json
// @Param input body userbalance.Transaction true "replenishment information"
// @Success 200 {object} userbalance.Response
// @Failure 500 {object} userbalance.Response
// @Router /topup [post]
func (h *Handler) replenishmentBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var transaction userbalance.Transaction

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := h.services.ReplenishmentBalance(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := &userbalance.Response{
		Message: "баланс пополнен",
	}
	json.NewEncoder(w).Encode(response)
}

// @Summary Money transfer
// @Tags balance
// @Description money transfer between users
// @ID transfer
// @Accept  json
// @Produce  json
// @Param input body userbalance.Money true "transfer information"
// @Success 200 {object} userbalance.Response
// @Failure 500 {object} userbalance.Response
// @Router /transfer [post]
func (h *Handler) transfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var money userbalance.Money

	if err := json.NewDecoder(r.Body).Decode(&money); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := h.services.Transfer(&money); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := &userbalance.Response{
		Message: "перевод стредств выполнен",
	}
	json.NewEncoder(w).Encode(response)
}

// @Summary Get History
// @Tags info
// @Description getting user history
// @ID get-history
// @Accept  json
// @Produce  json
// @Param input body userbalance.RequestHistory true "history request information"
// @Success 200 {object} []userbalance.History
// @Failure 500 {object} userbalance.Response
// @Router /history [post]
func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var requestHistory userbalance.RequestHistory
	var history []userbalance.History

	if err = json.NewDecoder(r.Body).Decode(&requestHistory); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if history, err = h.services.GetHistory(&requestHistory); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}

// @Summary Get Report
// @Tags info
// @Description getting report for the specified period
// @ID get-report
// @Accept  json
// @Produce  json
// @Param input body userbalance.RequestReport true "report request information"
// @Success 200 {object} []userbalance.Report
// @Failure 500 {object} userbalance.Response
// @Router /report [post]
func (h *Handler) createReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var requestReport userbalance.RequestReport
	var reportPath string

	if err = json.NewDecoder(r.Body).Decode(&requestReport); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if reportPath, err = h.services.CreateReport(&requestReport); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := &userbalance.Response{
		Message: reportPath,
	}
	json.NewEncoder(w).Encode(response)
}

// @Summary Reservation of funds
// @Tags balance
// @Description reservation of funds
// @ID reservation
// @Accept  json
// @Produce  json
// @Param input body userbalance.Transaction true "transaction info"
// @Success 200 {object} userbalance.Response
// @Failure 500 {object} userbalance.Response
// @Router /reserv [post]
func (h *Handler) reservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction userbalance.Transaction

	if err = json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if err = h.services.Reservation(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := &userbalance.Response{
		Message: "резервирование средств прошло успешно",
	}
	json.NewEncoder(w).Encode(response)
}

// @Summary Confirmation of funds
// @Tags balance
// @Description confirmation of funds
// @ID confirmation
// @Accept  json
// @Produce  json
// @Param input body userbalance.Transaction true "transaction info"
// @Success 200 {object} userbalance.Response
// @Failure 500 {object} userbalance.Response
// @Router /confirm [post]
func (h *Handler) confirmation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction userbalance.Transaction

	if err = json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if err = h.services.Confirmation(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := &userbalance.Response{
		Message: "средства из резерва были списаны успешно",
	}
	json.NewEncoder(w).Encode(response)
}

// @Summary Cancel Reservation
// @Tags balance
// @Description cancel reservation
// @ID cancel
// @Accept  json
// @Produce  json
// @Param input body userbalance.Transaction true "transaction info"
// @Success 200 {object} userbalance.Response
// @Failure 500 {object} userbalance.Response
// @Router /cancel [post]
func (h *Handler) cancelReservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction userbalance.Transaction

	if err = json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	if err = h.services.CancelReservation(&transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		response := &userbalance.Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := &userbalance.Response{
		Message: "разрезервирование средств прошло успешно",
	}
	json.NewEncoder(w).Encode(response)
}
