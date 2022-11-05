package repository

import (
	"database/sql"
	"time"
	"userbalance/internal/models"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type Repository struct {
	Control
}

type Control interface {
	UpdateBalanceTx(tx *sql.Tx, userId int, amount int) error
	GetUser(userId int) (*models.User, error)
	GetUserForUpdate(tx *sql.Tx, userId int) (*models.User, error)
	InsertUserTx(tx *sql.Tx, userId int, amount int) error
	InsertLogTx(tx *sql.Tx, userId int, date string, amount int, description string) error
	InsertMoneyReserveAccountsTx(tx *sql.Tx, userId int) error
	UpdateMoneyReserveAccountsTx(tx *sql.Tx, userId int, amount int) error
	GetBalanceReserveAccounts(userId int) (int, error)
	InsertMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int, date string) error
	DeleteMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int, date string) (int64, error)
	InsertReportTx(tx *sql.Tx, userId, serviceId, amount int, date string) error
	GetService(serviceId int) (string, error)
	GetReport(fromDate time.Time, toDate time.Time) (map[string]int, error)
	GetHistory(requestHistory *models.RequestHistory) ([]models.History, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Control: NewControlPostgres(db),
	}
}
