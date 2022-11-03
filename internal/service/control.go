package service

import (
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

type ControlService struct {
	repo repository.Control
}

func NewControlService(repo repository.Control) *ControlService {
	return &ControlService{repo: repo}
}

func (c *ControlService) GetBalance(userId int) (models.User, error) {
	var user models.User
	var err error

	if user, err = c.repo.GetBalance(userId); err != nil {
		return models.User{}, err
	}
	if user.Id == 0 {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, err
}

func (c *ControlService) ReplenishmentBalance(transaction *models.Transaction) error {
	if transaction.UserID <= 0 {
		return fmt.Errorf("не возможно создать пользователя с id = %d", transaction.UserID)
	}
	if transaction.Amount <= 0 {
		return fmt.Errorf("сумма пополнения должна быть больше 0")
	}
	date, _ := time.Parse("2006-01-02", transaction.Date)
	if date.Format("2006-01-02") == "0001-01-01" {
		transaction.Date = time.Now().Format("2006-01-02")

	} else {
		transaction.Date = date.Format("2006-01-02")
	}
	return c.repo.ReplenishmentBalance(transaction.UserID, transaction.Amount, transaction.Date)
}

func (c *ControlService) Transfer(money *models.Money) error {
	var user models.User
	var err error

	if money.Amount <= 0 {
		return errors.New("сумма перевода должна быть больше 0")
	}

	if money.FromUserID == money.ToUserID {
		return errors.New("невозможно перевести самому себе")
	}

	date, _ := time.Parse("2006-01-02", money.Date)
	if date.Format("2006-01-02") == "0001-01-01" {
		money.Date = time.Now().Format("2006-01-02")

	} else {
		money.Date = date.Format("2006-01-02")
	}

	if user, err = c.GetBalance(money.FromUserID); err != nil {
		return err
	}

	if user.Balance-money.Amount < 0 {
		return errors.New("недостаточно средств")
	} else {
		return c.repo.Transfer(money.FromUserID, money.ToUserID, money.Amount, money.Date)
	}
}

func (c *ControlService) Reservation(transaction *models.Transaction) error {
	var user models.User
	var err error

	if transaction.OrderID <= 0 {
		return errors.New("укажите номер заказа")
	}

	if transaction.Amount <= 0 {
		return errors.New("стоимость услуги не может быть меньше 0")
	}

	date, _ := time.Parse("2006-01-02", transaction.Date)
	if date.Format("2006-01-02") == "0001-01-01" {
		transaction.Date = time.Now().Format("2006-01-02")
	} else {
		transaction.Date = date.Format("2006-01-02")
	}

	if user, err = c.GetBalance(transaction.UserID); err != nil {
		return err
	}

	if user.Balance-transaction.Amount < 0 {
		return errors.New("недостаточно средств")
	} else {
		return c.repo.Reservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date)
	}
}

func (c *ControlService) CancelReservation(transaction *models.Transaction) error {

	date, _ := time.Parse("2006-01-02", transaction.Date)
	if date.Format("2006-01-02") == "0001-01-01" {
		transaction.Date = time.Now().Format("2006-01-02")
	} else {
		transaction.Date = date.Format("2006-01-02")
	}

	return c.repo.CancelReservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date)
}

func (c *ControlService) Confirmation(transaction *models.Transaction) error {

	date, _ := time.Parse("2006-01-02", transaction.Date)
	if date.Format("2006-01-02") == "0001-01-01" {
		transaction.Date = time.Now().Format("2006-01-02")
	} else {
		transaction.Date = date.Format("2006-01-02")
	}

	return c.repo.Confirmation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date)
}

func (c *ControlService) CreateReport(requestReport *models.RequestReport) (string, error) {
	var report map[string]int
	var err error
	var file *os.File
	var path string
	var dir string = "file"

	from, _ := time.Parse("2006-01-02", requestReport.FromDate)
	to, _ := time.Parse("2006-01-02", requestReport.ToDate)
	if (from.Format("2006-01-02") == "0001-01-01") || (to.Format("2006-01-02") == "0001-01-01") {
		requestReport.FromDate = utility.BeginningOfMonth().Format("2006-01-02")
		requestReport.ToDate = utility.EndOfMonth().Format("2006-01-02")
	} else {
		requestReport.FromDate = from.Format("2006-01-02")
		requestReport.ToDate = to.Format("2006-01-02")
	}

	if report, err = c.repo.CreateReport(requestReport.FromDate, requestReport.ToDate); err != nil {
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
