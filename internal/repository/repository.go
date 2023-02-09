package repository

import (
	"database/sql"
	"time"
	"userbalance/pkg/api"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type Repository struct {
	Control
}

type Control interface {
	UpdateBalanceTx(tx *sql.Tx, userId int32, amount int32) error
	GetUser(userId int32) (*api.User, error)
	GetUserForUpdate(tx *sql.Tx, userId int32) (*api.User, error)
	InsertUserTx(tx *sql.Tx, userId int32, amount int32) error
	InsertLogTx(tx *sql.Tx, userId int32, date time.Time, amount int32, description string) error
	InsertMoneyReserveAccountsTx(tx *sql.Tx, userId int32) error
	UpdateMoneyReserveAccountsTx(tx *sql.Tx, userId int32, amount int32) error
	GetBalanceReserveAccountsTx(tx *sql.Tx, userId int32) (int32, error)
	InsertMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int32, date time.Time) error
	DeleteMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int32, date time.Time) (int64, error)
	InsertReportTx(tx *sql.Tx, userId, serviceId, amount int32, date time.Time) error
	GetService(serviceId int32) (string, error)
	GetReport(fromDate time.Time, toDate time.Time) (map[string]int32, error)
	GetHistory(requestHistory *api.RequestHistory) ([]*api.History, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Control: NewControlPostgres(db),
	}
}
