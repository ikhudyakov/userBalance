package handler

import (
	"log"
	"net/http"
	"userbalance/internal/models"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/mailru/easyjson"
)

// @Summary Get Balance
// @Tags balance
// @Description getting the user's balance
// @ID get-balance
// @Accept  json
// @Produce  json
// @Param input body models.User true "user id"
// @Success 200 {object} models.User
// @Failure 500 {object} models.Response
// @Router / [post]
func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var user models.User
	var newUser *models.User

	if err = easyjson.UnmarshalFromReader(r.Body, &user); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.Validate(user.Id,
		validation.Required.Error("id пользователя не может быть <= 0"),
		validation.Min(1).Error("id пользователя не может быть <= 0"),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if newUser, err = h.services.GetBalance(user.Id); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(newUser, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Replenishment Balance
// @Tags balance
// @Description replenishment of the user's balance
// @ID replenishment-balance
// @Accept  json
// @Produce  json
// @Param input body models.Transaction true "replenishment information"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /topup [post]
func (h *Handler) replenishmentBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction models.Transaction

	if err = easyjson.UnmarshalFromReader(r.Body, &transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.ValidateStruct(&transaction,
		validation.Field(&transaction.UserID, validation.Required.Error("id пользователя не может быть <= 0"), validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&transaction.Amount, validation.Required.Error("сумма пополнения должна быть больше 0"), validation.Min(1).Error("сумма пополнения должна быть больше 0")),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if err = h.services.ReplenishmentBalance(&transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	response := &models.Response{
		Message: "баланс пополнен",
	}
	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(response, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Money transfer
// @Tags balance
// @Description money transfer between users
// @ID transfer
// @Accept  json
// @Produce  json
// @Param input body models.Money true "transfer information"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /transfer [post]
func (h *Handler) transfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var money models.Money

	if err = easyjson.UnmarshalFromReader(r.Body, &money); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.ValidateStruct(&money,
		validation.Field(&money.FromUserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&money.ToUserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0"),
			validation.NotIn(money.FromUserID).Error("невозможно перевести самому себе")),
		validation.Field(&money.Amount,
			validation.Required.Error("сумма перевода должна быть больше 0"),
			validation.Min(1).Error("сумма перевода должна быть больше 0")),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if err := h.services.Transfer(&money); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	response := &models.Response{
		Message: "перевод стредств выполнен",
	}
	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(response, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Get History
// @Tags info
// @Description getting user history
// @ID get-history
// @Accept  json
// @Produce  json
// @Param input body models.RequestHistory true "history request information"
// @Success 200 {object} []models.History
// @Failure 500 {object} models.Response
// @Router /history [post]
func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var requestHistory models.RequestHistory
	var history []models.History

	if err = easyjson.UnmarshalFromReader(r.Body, &requestHistory); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.Validate(requestHistory.UserID,
		validation.Required.Error("id пользователя не может быть <= 0"),
		validation.Min(1).Error("id пользователя не может быть <= 0"),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if history, err = h.services.GetHistory(&requestHistory); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(&models.Histories{
		Entity: history,
	}, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Get Report
// @Tags info
// @Description getting report for the specified period
// @ID get-report
// @Accept  json
// @Produce  json
// @Param input body models.RequestReport true "report request information"
// @Success 200 {object} []models.Report
// @Failure 500 {object} models.Response
// @Router /report [post]
func (h *Handler) createReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var requestReport models.RequestReport
	var reportPath string

	if err = easyjson.UnmarshalFromReader(r.Body, &requestReport); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if reportPath, err = h.services.CreateReport(&requestReport); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	response := &models.Response{
		Message: reportPath,
	}
	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(response, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Reservation of funds
// @Tags balance
// @Description reservation of funds
// @ID reservation
// @Accept  json
// @Produce  json
// @Param input body models.Transaction true "transaction info"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /reserv [post]
func (h *Handler) reservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction models.Transaction

	if err = easyjson.UnmarshalFromReader(r.Body, &transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.ValidateStruct(&transaction,
		validation.Field(&transaction.UserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&transaction.Amount,
			validation.Required.Error("стоимость услуги должна быть больше 0"),
			validation.Min(1).Error("стоимость услуги должна быть больше 0")),
		validation.Field(&transaction.OrderID,
			validation.Required.Error("номер заказа не может быть <= 0"),
			validation.Min(1).Error("номер заказа не может быть <= 0")),
		validation.Field(&transaction.ServiceID,
			validation.Required.Error("id услуги не может быть <= 0"),
			validation.Min(1).Error("id услуги не может быть <= 0")),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if err = h.services.Reservation(&transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	response := &models.Response{
		Message: "резервирование средств прошло успешно",
	}
	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(response, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Confirmation of funds
// @Tags balance
// @Description confirmation of funds
// @ID confirmation
// @Accept  json
// @Produce  json
// @Param input body models.Transaction true "transaction info"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /confirm [post]
func (h *Handler) confirmation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction models.Transaction

	if err = easyjson.UnmarshalFromReader(r.Body, &transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.ValidateStruct(&transaction,
		validation.Field(&transaction.UserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&transaction.Amount,
			validation.Required.Error("стоимость услуги должна быть больше 0"),
			validation.Min(1).Error("стоимость услуги должна быть больше 0")),
		validation.Field(&transaction.OrderID,
			validation.Required.Error("номер заказа не может быть <= 0"),
			validation.Min(1).Error("номер заказа не может быть <= 0")),
		validation.Field(&transaction.ServiceID,
			validation.Required.Error("id услуги не может быть <= 0"),
			validation.Min(1).Error("id услуги не может быть <= 0")),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if err = h.services.Confirmation(&transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	response := &models.Response{
		Message: "средства из резерва были списаны успешно",
	}
	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(response, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

// @Summary Cancel Reservation
// @Tags balance
// @Description cancel reservation
// @ID cancel
// @Accept  json
// @Produce  json
// @Param input body models.Transaction true "transaction info"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /cancel [post]
func (h *Handler) cancelReservation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var transaction models.Transaction

	if err = easyjson.UnmarshalFromReader(r.Body, &transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	if err = validation.ValidateStruct(&transaction,
		validation.Field(&transaction.UserID,
			validation.Required.Error("id пользователя не может быть <= 0"),
			validation.Min(1).Error("id пользователя не может быть <= 0")),
		validation.Field(&transaction.Amount,
			validation.Required.Error("стоимость услуги должна быть больше 0"),
			validation.Min(1).Error("стоимость услуги должна быть больше 0")),
		validation.Field(&transaction.OrderID,
			validation.Required.Error("номер заказа не может быть <= 0"),
			validation.Min(1).Error("номер заказа не может быть <= 0")),
		validation.Field(&transaction.ServiceID,
			validation.Required.Error("id услуги не может быть <= 0"),
			validation.Min(1).Error("id услуги не может быть <= 0")),
	); err != nil {
		Error(err, w, http.StatusBadRequest)
		return
	}

	if err = h.services.CancelReservation(&transaction); err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}

	response := &models.Response{
		Message: "разрезервирование средств прошло успешно",
	}
	w.WriteHeader(http.StatusOK)
	_, err = easyjson.MarshalToWriter(response, w)
	if err != nil {
		Error(err, w, http.StatusInternalServerError)
		return
	}
}

func Error(err error, w http.ResponseWriter, status int) {
	log.Println(err.Error())
	response := &models.Response{
		Message: err.Error(),
	}
	res, err := easyjson.Marshal(response)
	if err != nil {
		log.Println(err.Error())
		return
	}
	w.WriteHeader(status)
	w.Write(res)
}
