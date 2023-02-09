package service

import (
	"database/sql"
	c "userbalance/internal/config"
	"userbalance/internal/repository"
	"userbalance/pkg/api"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Control interface {
	ReplenishmentBalance(replenishment *api.Replenishment) error
	Transfer(money *api.Money) error
	Reservation(transaction *api.Transaction) error
	CancelReservation(transaction *api.Transaction) error
	Confirmation(transaction *api.Transaction) error
	GetBalance(userId int32) (*api.User, error)
	CreateReport(requestReport *api.RequestReport) (string, error)
	GetHistory(requestHistory *api.RequestHistory) ([]*api.History, error)
}

type Service struct {
	Control
}

func NewService(repos *repository.Repository, conf *c.Config, db *sql.DB) *Service {
	return &Service{
		Control: NewControlService(repos.Control, conf, db),
	}
}
