package handler

import (
	"context"
	"log"
	"userbalance/pkg/api"
)

func (h *Handler) GetBalance(ctx context.Context, user *api.User) (*api.User, error) {
	var err error
	var newUser *api.User

	if err = user.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if newUser, err = h.services.GetBalance(user.GetId()); err != nil {
		log.Println(err)
		return nil, err
	}

	log.Printf("user id = %d, balance = %d", newUser.Id, newUser.Balance)
	return newUser, nil
}

func (h *Handler) ReplenishmentBalance(ctx context.Context, replenishment *api.Replenishment) (*api.Response, error) {
	var err error

	if err = replenishment.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err = h.services.ReplenishmentBalance(replenishment); err != nil {
		log.Println(err)
		return nil, err
	}

	response := &api.Response{
		Message: "баланс пополнен",
	}

	return response, nil
}

func (h *Handler) Transfer(ctx context.Context, money *api.Money) (*api.Response, error) {
	var err error

	if err = money.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := h.services.Transfer(money); err != nil {
		log.Println(err)
		return nil, err
	}

	response := &api.Response{
		Message: "перевод стредств выполнен",
	}

	return response, nil
}

func (h *Handler) GetHistory(ctx context.Context, requestHistory *api.RequestHistory) (*api.Histories, error) {
	var err error
	var history []*api.History

	if err = requestHistory.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if history, err = h.services.GetHistory(requestHistory); err != nil {
		log.Println(err)
		return nil, err
	}

	return &api.Histories{Entity: history}, nil
}

func (h *Handler) Reservation(ctx context.Context, transaction *api.Transaction) (*api.Response, error) {
	var err error

	if err = transaction.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err = h.services.Reservation(transaction); err != nil {
		log.Println(err)
		return nil, err
	}

	response := &api.Response{
		Message: "резервирование средств прошло успешно",
	}

	return response, nil
}

func (h *Handler) Confirmation(ctx context.Context, transaction *api.Transaction) (*api.Response, error) {
	var err error

	if err = transaction.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err = h.services.Confirmation(transaction); err != nil {
		log.Println(err)
		return nil, err
	}

	response := &api.Response{
		Message: "средства из резерва были списаны успешно",
	}

	return response, nil
}

func (h *Handler) CancelReservation(ctx context.Context, transaction *api.Transaction) (*api.Response, error) {
	var err error

	if err = transaction.Validate(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err = h.services.CancelReservation(transaction); err != nil {
		log.Println(err)
		return nil, err
	}

	response := &api.Response{
		Message: "разрезервирование средств прошло успешно",
	}

	return response, nil
}

// func (h *Handler) createReport(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	var err error
// 	var requestReport models.RequestReport
// 	var reportPath string

// 	if err = easyjson.UnmarshalFromReader(r.Body, &requestReport); err != nil {
// 		Error(err, w, http.StatusInternalServerError)
// 		return
// 	}

// 	if err = requestReport.Validate(); err != nil {
// 		Error(err, w, http.StatusBadRequest)
// 		return
// 	}

// 	if reportPath, err = h.services.CreateReport(&requestReport); err != nil {
// 		Error(err, w, http.StatusInternalServerError)
// 		return
// 	}

// 	response := &models.Response{
// 		Message: reportPath,
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	_, err = easyjson.MarshalToWriter(response, w)
// 	if err != nil {
// 		Error(err, w, http.StatusInternalServerError)
// 		return
// 	}
// }
