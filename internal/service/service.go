package service

import (
	"userbalance"
	"userbalance/internal/repository"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Control interface {
	ReplenishmentBalance(transaction *userbalance.Transaction) error
	Transfer(money *userbalance.Money) error
	Reservation(transaction *userbalance.Transaction) error
	CancelReservation(transaction *userbalance.Transaction) error
	Confirmation(transaction *userbalance.Transaction) error
	GetBalance(userId int) (userbalance.User, error)
	CreateReport(requestReport *userbalance.RequestReport) (string, error)
	GetHistory(requestHistory *userbalance.RequestHistory) ([]userbalance.History, error)
}

type Service struct {
	Control
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Control: NewControlService(repos.Control),
	}
}
