package service

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	c "userbalance/internal/config"
	"userbalance/internal/repository"
	"userbalance/pkg/api"
)

const layout string = "2006-01-02"

type ControlService struct {
	repo repository.Control
	conf *c.Config
	db   *sql.DB
}

func NewControlService(repo repository.Control, conf *c.Config, db *sql.DB) *ControlService {
	return &ControlService{
		repo: repo,
		conf: conf,
		db:   db,
	}
}

func (c *ControlService) GetBalance(userId int32) (*api.User, error) {
	var user *api.User
	var err error

	if user, err = c.repo.GetUser(userId); err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("пользователь не найден")
	}
	return user, err
}

func (c *ControlService) ReplenishmentBalance(replenishment *api.Replenishment) error {
	var tx *sql.Tx
	var err error
	var user *api.User

	date, _ := time.Parse(layout, replenishment.Date)
	if date.IsZero() {
		date = time.Now()
	}

	tx, err = c.db.Begin()
	if err != nil {
		return err
	}

	user, err = c.repo.GetUserForUpdate(tx, replenishment.UserID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if user != nil {
		if err = c.repo.UpdateBalanceTx(tx, replenishment.UserID, user.Balance+replenishment.Amount); err != nil {
			tx.Rollback()
			return err
		}
		if err = c.repo.InsertLogTx(tx, replenishment.UserID, date, replenishment.Amount, "Пополнение баланса"); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err = c.repo.InsertUserTx(tx, replenishment.UserID, replenishment.Amount); err != nil {
			tx.Rollback()
			return err
		}
		if err = c.repo.InsertMoneyReserveAccountsTx(tx, replenishment.UserID); err != nil {
			tx.Rollback()
			return err
		}
		if err = c.repo.InsertLogTx(tx, replenishment.UserID, date, replenishment.Amount, "Пополнение баланса"); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (c *ControlService) Transfer(money *api.Money) error {
	var tx *sql.Tx
	var err error
	var fromUser, toUser *api.User

	date, _ := time.Parse(layout, money.Date)
	if date.IsZero() {
		date = time.Now()
	}

	tx, err = c.db.Begin()
	if err != nil {
		return err
	}

	if fromUser, err = c.repo.GetUserForUpdate(tx, money.FromUserID); err != nil {
		tx.Rollback()
		return err
	}
	if fromUser == nil {
		tx.Rollback()
		return errors.New("пользователь не найден")
	}
	if toUser, err = c.repo.GetUserForUpdate(tx, money.ToUserID); err != nil {
		tx.Rollback()
		return err
	}
	if toUser == nil {
		tx.Rollback()
		return errors.New("пользователь не найден")
	}

	if fromUser.Balance-money.Amount < 0 {
		tx.Rollback()
		return errors.New("недостаточно средств")
	}

	if err = c.repo.UpdateBalanceTx(tx, fromUser.Id, fromUser.Balance-money.Amount); err != nil {
		tx.Rollback()
		return err
	}
	if err = c.repo.InsertLogTx(tx, money.FromUserID, date, money.Amount, fmt.Sprintf("Перевод средств пользователю %d", money.ToUserID)); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.UpdateBalanceTx(tx, toUser.Id, toUser.Balance+money.Amount); err != nil {
		tx.Rollback()
		return err
	}
	if err = c.repo.InsertLogTx(tx, money.ToUserID, date, money.Amount, fmt.Sprintf("Перевод средств от пользователя %d", money.FromUserID)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) Reservation(transaction *api.Transaction) error {
	var tx *sql.Tx
	var user *api.User
	var service string
	var err error
	var reservBalance int32

	date, _ := time.Parse(layout, transaction.Date)
	if date.IsZero() {
		date = time.Now()
	}

	if service, err = c.repo.GetService(transaction.ServiceID); err != nil {
		return err
	}

	if service == "" {
		return errors.New("услуга не найдена")
	}

	tx, err = c.db.Begin()
	if err != nil {
		return err
	}

	if user, err = c.repo.GetUserForUpdate(tx, transaction.UserID); err != nil {
		tx.Rollback()
		return err
	}
	if user == nil {
		tx.Rollback()
		return errors.New("пользователь не найден")
	}
	if user.Balance-transaction.Amount < 0 {
		tx.Rollback()
		return errors.New("недостаточно средств")
	}

	if reservBalance, err = c.repo.GetBalanceReserveAccountsTx(tx, transaction.UserID); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.UpdateBalanceTx(tx, transaction.UserID, user.Balance-transaction.Amount); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.UpdateMoneyReserveAccountsTx(tx, transaction.UserID, reservBalance+transaction.Amount); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.InsertMoneyReserveDetailsTx(tx, transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, date); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.InsertLogTx(tx, transaction.UserID, date, transaction.Amount, fmt.Sprintf("Заказ №%d, услуга \"%s\"", transaction.OrderID, service)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) CancelReservation(transaction *api.Transaction) error {
	var tx *sql.Tx
	var user *api.User
	var service string
	var err error
	var reservBalance int32

	date, _ := time.Parse(layout, transaction.Date)
	if date.IsZero() {
		date = time.Now()
	}

	if service, err = c.repo.GetService(transaction.ServiceID); err != nil {
		return err
	}

	if service == "" {
		return errors.New("услуга не найдена")
	}

	tx, err = c.db.Begin()
	if err != nil {
		return err
	}

	if user, err = c.repo.GetUserForUpdate(tx, transaction.UserID); err != nil {
		tx.Rollback()
		return err
	}
	if user == nil {
		tx.Rollback()
		return errors.New("пользователь не найден")
	}

	if reservBalance, err = c.repo.GetBalanceReserveAccountsTx(tx, transaction.UserID); err != nil {
		tx.Rollback()
		return err
	}

	r, err := c.repo.DeleteMoneyReserveDetailsTx(tx, transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, date)
	if err != nil {
		tx.Rollback()
		return err
	}
	if r == 0 {
		tx.Rollback()
		return errors.New("по указанным критериям не было резерва")
	}

	if err = c.repo.InsertLogTx(tx, transaction.UserID, date, transaction.Amount, fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", transaction.OrderID, service)); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.UpdateBalanceTx(tx, transaction.UserID, user.Balance+transaction.Amount); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.UpdateMoneyReserveAccountsTx(tx, transaction.UserID, reservBalance-transaction.Amount); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) Confirmation(transaction *api.Transaction) error {
	var tx *sql.Tx
	var err error
	var reservBalance int32

	date, _ := time.Parse(layout, transaction.Date)
	if date.IsZero() {
		date = time.Now()
	}

	tx, err = c.db.Begin()
	if err != nil {
		return err
	}

	if reservBalance, err = c.repo.GetBalanceReserveAccountsTx(tx, transaction.UserID); err != nil {
		tx.Rollback()
		return err
	}

	r, err := c.repo.DeleteMoneyReserveDetailsTx(tx, transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, date)
	if err != nil {
		tx.Rollback()
		return err
	}
	if r == 0 {
		tx.Rollback()
		return errors.New("по указанным критериям не было резерва")
	}

	if err = c.repo.UpdateMoneyReserveAccountsTx(tx, transaction.UserID, reservBalance-transaction.Amount); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.InsertReportTx(tx, transaction.UserID, transaction.ServiceID, transaction.Amount, date); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) CreateReport(requestReport *api.RequestReport) (string, error) {
	var report map[string]int32
	var err error
	var file *os.File
	var path string
	var dir string = "file"

	from := time.Date(int(requestReport.Year), time.Month(requestReport.Month), 1, 0, 0, 0, 0, time.Local)
	to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)

	if report, err = c.repo.GetReport(from, to); err != nil {
		return path, err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0777); err != nil {
			return path, err
		}
	}

	if file, err = os.Create(fmt.Sprintf("%s/%d.csv", dir, time.Now().Unix())); err != nil {
		return path, err
	}
	defer file.Close()

	for k, v := range report {
		_, err = file.WriteString(fmt.Sprintf("%s;%v\n", k, v))

		if err != nil {
			return path, err
		}
	}

	var host string

	if c.conf.Host == "" {
		host = "localhost"
	} else {
		host = c.conf.Host
	}

	if c.conf.Port != "" {
		host += c.conf.Port
	}

	path = fmt.Sprintf("%s/%s", host, file.Name())

	return path, err
}

func (c *ControlService) GetHistory(requestHistory *api.RequestHistory) ([]*api.History, error) {
	direction := strings.ToUpper(requestHistory.Direction)
	sortField := strings.ToLower(requestHistory.SortField)

	if direction == "DESC" {
		requestHistory.Direction = "DESC"
	} else {
		requestHistory.Direction = "ASC"
	}

	if sortField == "date" {
		requestHistory.SortField = "date"
	} else {
		requestHistory.SortField = "amount"
	}

	history, err := c.repo.GetHistory(requestHistory)
	if err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return history, errors.New("записи не найдены")
	}
	return history, err
}
