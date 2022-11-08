package service

import (
	"database/sql"
	c "userbalance/internal/config"
	"userbalance/internal/models"
	"userbalance/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Control interface {
	ReplenishmentBalance(replenishment *models.Replenishment) error
	Transfer(money *models.Money) error
	Reservation(transaction *models.Transaction) error
	CancelReservation(transaction *models.Transaction) error
	Confirmation(transaction *models.Transaction) error
	GetBalance(userId int) (*models.User, error)
	CreateReport(requestReport *models.RequestReport) (string, error)
	GetHistory(requestHistory *models.RequestHistory) ([]models.History, error)
}

type Service struct {
	Control
}

func NewService(repos *repository.Repository, conf *c.Config, db *sql.DB) *Service {
	return &Service{
		Control: NewControlService(repos.Control, conf, db),
	}
}
