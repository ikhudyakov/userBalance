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
	UpdateBalanceTx(tx *sql.Tx, userId int, amount int) error
	GetUser(userId int) (*models.User, error)
	InsertUserTx(tx *sql.Tx, userId int, amount int) error
	InsertLogTx(tx *sql.Tx, userId int, date string, amount int, description string) error
	InsertMoneyReserveAccountsTx(tx *sql.Tx, userId int) error
	UpdateMoneyReserveAccountsTx(tx *sql.Tx, userId int, amount int) error
	GetBalanceReserveAccounts(userId int) (int, error)
	InsertMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int, date string) error
	DeleteMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int, date string) (int64, error)
	InsertReportTx(tx *sql.Tx, userId, serviceId, amount int, date string) error
	GetService(serviceId int) (string, error)
	GetReport(fromDate string, toDate string) (map[string]int, error)
	GetHistory(userId int) ([]models.History, error)
	GetDB() *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Control: NewControlPostgres(db),
	}
}
