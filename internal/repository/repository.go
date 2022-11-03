package repository

import (
	"database/sql"
	"userbalance/internal/models"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type Repository struct {
	Control
}

type Control interface {
	ReplenishmentBalance(userId int, amount int, date string) error
	Transfer(fromUserId int, toUserId int, amount int, date string) error
	Reservation(userId int, serviceId int, orderId int, amount int, date string) error
	Confirmation(userId int, serviceId int, orderId int, amount int, date string) error
	CancelReservation(userId int, serviceId int, orderId int, amount int, date string) error
	GetBalance(userId int) (models.User, error)
	CreateReport(fromDate string, toDate string) (map[string]int, error)
	GetHistory(userId int) ([]models.History, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Control: NewControlPostgres(db),
	}
}
