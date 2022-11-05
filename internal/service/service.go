package service

import (
	"userbalance/internal/models"
	"userbalance/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Control interface {
	ReplenishmentBalance(transaction *models.Transaction) error
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

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Control: NewControlService(repos.Control),
	}
}
