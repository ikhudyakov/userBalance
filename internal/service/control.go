package service

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"userbalance/internal/models"
	"userbalance/internal/repository"
	"userbalance/internal/utility"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const layout string = "2006-01-02"

type ControlService struct {
	repo repository.Control
}

func NewControlService(repo repository.Control) *ControlService {
	return &ControlService{repo: repo}
}

func (c *ControlService) GetBalance(userId int) (*models.User, error) {
	var user *models.User
	var err error

	if user, err = c.repo.GetUser(userId); err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("пользователь не найден")
	}
	return user, err
}

func (c *ControlService) ReplenishmentBalance(transaction *models.Transaction) error {
	var tx *sql.Tx
	var err error
	var user *models.User

	date, _ := time.Parse(layout, transaction.Date)

	if time.Time.IsZero(date) {
		transaction.Date = time.Now().Format(layout)

	} else {
		transaction.Date = date.Format(layout)
	}

	user, err = c.repo.GetUser(transaction.UserID)
	if err != nil {
		return err
	}

	tx, err = c.repo.GetDB().Begin()
	if err != nil {
		return err
	}

	if user != nil {
		if err = c.repo.UpdateBalanceTx(tx, transaction.UserID, user.Balance+transaction.Amount); err != nil {
			tx.Rollback()
			return err
		}
		if err = c.repo.InsertLogTx(tx, transaction.UserID, transaction.Date, transaction.Amount, "Пополнение баланса"); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err = c.repo.InsertUserTx(tx, transaction.UserID, transaction.Amount); err != nil {
			tx.Rollback()
			return err
		}
		if err = c.repo.InsertMoneyReserveAccountsTx(tx, transaction.UserID); err != nil {
			tx.Rollback()
			return err
		}
		if err = c.repo.InsertLogTx(tx, transaction.UserID, transaction.Date, transaction.Amount, "Пополнение баланса"); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (c *ControlService) Transfer(money *models.Money) error {
	var tx *sql.Tx
	var err error
	var fromUser, toUser *models.User

	date, _ := time.Parse(layout, money.Date)
	if time.Time.IsZero(date) {
		money.Date = time.Now().Format(layout)

	} else {
		money.Date = date.Format(layout)
	}

	if fromUser, err = c.GetBalance(money.FromUserID); err != nil {
		return err
	}
	if toUser, err = c.GetBalance(money.ToUserID); err != nil {
		return err
	}

	if fromUser.Balance-money.Amount < 0 {
		return errors.New("недостаточно средств")
	}

	tx, err = c.repo.GetDB().Begin()
	if err != nil {
		return err
	}

	if err = c.repo.UpdateBalanceTx(tx, fromUser.Id, fromUser.Balance-money.Amount); err != nil {
		tx.Rollback()
		return err
	}
	if err = c.repo.InsertLogTx(tx, money.FromUserID, money.Date, money.Amount, fmt.Sprintf("Перевод средств пользователю %d", money.ToUserID)); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.UpdateBalanceTx(tx, toUser.Id, toUser.Balance+money.Amount); err != nil {
		tx.Rollback()
		return err
	}
	if err = c.repo.InsertLogTx(tx, money.ToUserID, money.Date, money.Amount, fmt.Sprintf("Перевод средств от пользователя %d", money.FromUserID)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) Reservation(transaction *models.Transaction) error {
	var tx *sql.Tx
	var user *models.User
	var service string
	var err error
	var reservBalance int

	date, _ := time.Parse(layout, transaction.Date)
	if time.Time.IsZero(date) {
		transaction.Date = time.Now().Format(layout)
	} else {
		transaction.Date = date.Format(layout)
	}

	if user, err = c.GetBalance(transaction.UserID); err != nil {
		return err
	}

	if user.Balance-transaction.Amount < 0 {
		return errors.New("недостаточно средств")
	}

	if service, err = c.repo.GetService(transaction.ServiceID); err != nil {
		return err
	}

	if service == "" {
		return errors.New("услуга не найдена")
	}

	if reservBalance, err = c.repo.GetBalanceReserveAccounts(transaction.UserID); err != nil {
		return err
	}

	tx, err = c.repo.GetDB().Begin()
	if err != nil {
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

	if err = c.repo.InsertMoneyReserveDetailsTx(tx, transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date); err != nil {
		tx.Rollback()
		return err
	}

	if err = c.repo.InsertLogTx(tx, transaction.UserID, transaction.Date, transaction.Amount, fmt.Sprintf("Заказ №%d, услуга \"%s\"", transaction.OrderID, service)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) CancelReservation(transaction *models.Transaction) error {
	var tx *sql.Tx
	var user *models.User
	var service string
	var err error
	var reservBalance int

	date, _ := time.Parse(layout, transaction.Date)
	if time.Time.IsZero(date) {
		transaction.Date = time.Now().Format(layout)
	} else {
		transaction.Date = date.Format(layout)
	}

	if user, err = c.GetBalance(transaction.UserID); err != nil {
		return err
	}

	if service, err = c.repo.GetService(transaction.ServiceID); err != nil {
		return err
	}

	if service == "" {
		return errors.New("услуга не найдена")
	}

	if reservBalance, err = c.repo.GetBalanceReserveAccounts(transaction.UserID); err != nil {
		return err
	}

	tx, err = c.repo.GetDB().Begin()
	if err != nil {
		return err
	}

	r, err := c.repo.DeleteMoneyReserveDetailsTx(tx, transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date)
	if err != nil {
		tx.Rollback()
		return err
	}
	if r == 0 {
		tx.Rollback()
		return errors.New("по указанным критериям не было резерва")
	}

	if err = c.repo.InsertLogTx(tx, transaction.UserID, transaction.Date, transaction.Amount, fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", transaction.OrderID, service)); err != nil {
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

func (c *ControlService) Confirmation(transaction *models.Transaction) error {
	var tx *sql.Tx
	var err error
	var reservBalance int

	date, _ := time.Parse(layout, transaction.Date)
	if time.Time.IsZero(date) {
		transaction.Date = time.Now().Format(layout)
	} else {
		transaction.Date = date.Format(layout)
	}

	if reservBalance, err = c.repo.GetBalanceReserveAccounts(transaction.UserID); err != nil {
		return err
	}

	tx, err = c.repo.GetDB().Begin()
	if err != nil {
		return err
	}

	r, err := c.repo.DeleteMoneyReserveDetailsTx(tx, transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date)
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

	if err = c.repo.InsertReportTx(tx, transaction.UserID, transaction.ServiceID, transaction.Amount, transaction.Date); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *ControlService) CreateReport(requestReport *models.RequestReport) (string, error) {
	var report map[string]int
	var err error
	var file *os.File
	var path string
	var dir string = "file"

	from, _ := time.Parse(layout, requestReport.FromDate)
	to, _ := time.Parse(layout, requestReport.ToDate)
	if time.Time.IsZero(from) || time.Time.IsZero(to) {
		requestReport.FromDate = utility.BeginningOfMonth().Format(layout)
		requestReport.ToDate = utility.EndOfMonth().Format(layout)
	} else {
		requestReport.FromDate = from.Format(layout)
		requestReport.ToDate = to.Format(layout)
	}

	if report, err = c.repo.GetReport(requestReport.FromDate, requestReport.ToDate); err != nil {
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

	tr := charmap.Windows1251.NewEncoder().Transformer
	encWriter := transform.NewWriter(file, tr)

	csvw := csv.NewWriter(encWriter)

	for k, v := range report {
		s := make([]string, 0)
		s = append(s, fmt.Sprintf("%s;%v", k, v))
		err = csvw.Write(s)
		if err != nil {
			return path, err
		}
	}
	defer csvw.Flush()

	path = "localhost:8081/" + file.Name()

	return path, err
}

func (c *ControlService) GetHistory(requestHistory *models.RequestHistory) ([]models.History, error) {
	history, err := c.repo.GetHistory(requestHistory.UserID)
	if len(history) == 0 {
		return history, errors.New("записи не найдены")
	}
	return Sort(history, requestHistory.SortField, requestHistory.Direction), err
}

func Sort(history []models.History, sortField, direction string) []models.History {
	if strings.ToLower(direction) == "desc" {
		if strings.ToLower(sortField) == "amount" {
			sort.SliceStable(history, func(i, j int) bool {
				return history[i].Amount > history[j].Amount
			})
		} else {
			sort.SliceStable(history, func(i, j int) bool {
				return history[i].Date > history[j].Date
			})
		}
	} else {
		if strings.ToLower(sortField) == "amount" {
			sort.SliceStable(history, func(i, j int) bool {
				return history[i].Amount < history[j].Amount
			})
		} else {
			sort.SliceStable(history, func(i, j int) bool {
				return history[i].Date < history[j].Date
			})
		}
	}
	return history
}
