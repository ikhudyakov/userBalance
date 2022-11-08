package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
	"userbalance/internal/config"
	"userbalance/internal/models"
	"userbalance/internal/repository"
	mock_repository "userbalance/internal/repository/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {

	type mockBehavior func(s *mock_repository.MockControl, userId int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		userId       int
		want         *models.User
		wantErr      bool
	}{
		{
			name:   "OK",
			userId: 1,
			mockBehavior: func(r *mock_repository.MockControl, userId int) {
				r.EXPECT().GetUser(userId).Return(
					&models.User{
						Id:      1,
						Balance: 100}, nil)
			},
			want: &models.User{
				Id:      1,
				Balance: 100,
			},
		},

		{
			name:    "error user not found",
			userId:  0,
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, userId int) {
				r.EXPECT().GetUser(userId).Return(
					nil, errors.New("пользователь не найден"))
			},
		},

		{
			name:    "error database",
			userId:  0,
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, userId int) {
				r.EXPECT().GetUser(userId).Return(
					nil, errors.New("error database"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.userId)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, nil)

			got, err := s.GetBalance(testCase.userId)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestReplenishmentBalance(t *testing.T) {

	type mockBehavior func(s *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User)

	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		replenishment *models.Replenishment
		user          *models.User
		date          time.Time
		wantErr       bool
	}{
		{
			name: "OK user exists",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			date: time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			user: &models.User{
				Id:      1,
				Balance: 200,
			},
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(user, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), replenishment.UserID, user.Balance+replenishment.Amount).Return(nil)
				r.EXPECT().InsertLogTx(gomock.Any(), replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "OK user not found",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			date: time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, nil)
				r.EXPECT().InsertUserTx(gomock.Any(), replenishment.UserID, replenishment.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveAccountsTx(gomock.Any(), replenishment.UserID).Return(nil)
				r.EXPECT().InsertLogTx(gomock.Any(), replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "error getuser",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalance",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 200,
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(user, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), replenishment.UserID, user.Balance+replenishment.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlog",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 200,
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(user, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), replenishment.UserID, user.Balance+replenishment.Amount).Return(nil)
				r.EXPECT().InsertLogTx(gomock.Any(), replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertuser",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, nil)
				r.EXPECT().InsertUserTx(gomock.Any(), replenishment.UserID, replenishment.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertmoneyreservacounts",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, nil)
				r.EXPECT().InsertUserTx(gomock.Any(), replenishment.UserID, replenishment.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveAccountsTx(gomock.Any(), replenishment.UserID).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlog",
			replenishment: &models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *models.Replenishment, user *models.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, nil)
				r.EXPECT().InsertUserTx(gomock.Any(), replenishment.UserID, replenishment.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveAccountsTx(gomock.Any(), replenishment.UserID).Return(nil)
				r.EXPECT().InsertLogTx(gomock.Any(), replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)

			testCase.mockBehavior(control, testCase.replenishment, testCase.user)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, db)

			err = s.ReplenishmentBalance(testCase.replenishment)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	type mockBehavior func(s *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		money        *models.Money
		fromUser     *models.User
		toUser       *models.User
		wantErr      bool
	}{
		{
			name: "OK",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.FromUserID, fromUser.Balance-money.Amount).Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					money.FromUserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					money.Amount,
					fmt.Sprintf("Перевод средств пользователю %d", money.ToUserID)).
					Return(nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.ToUserID, toUser.Balance+money.Amount).Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					money.ToUserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					money.Amount,
					fmt.Sprintf("Перевод средств от пользователя %d", money.FromUserID)).
					Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "error fromuser not found",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error touser not found",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error insufficient funds",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     2000,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error getfromuser",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     2000,
				Date:       "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error gettouser",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     2000,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalancefromuser",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.FromUserID, fromUser.Balance-money.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlogfromuser",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.FromUserID, fromUser.Balance-money.Amount).Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					money.FromUserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					money.Amount,
					fmt.Sprintf("Перевод средств пользователю %d", money.ToUserID)).
					Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalancetouser",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.FromUserID, fromUser.Balance-money.Amount).Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					money.FromUserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					money.Amount,
					fmt.Sprintf("Перевод средств пользователю %d", money.ToUserID)).
					Return(nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.ToUserID, toUser.Balance+money.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlogtouser",
			money: &models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &models.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &models.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *models.User, toUser *models.User, money *models.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.FromUserID, fromUser.Balance-money.Amount).Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					money.FromUserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					money.Amount,
					fmt.Sprintf("Перевод средств пользователю %d", money.ToUserID)).
					Return(nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.ToUserID, toUser.Balance+money.Amount).Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					money.ToUserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					money.Amount,
					fmt.Sprintf("Перевод средств от пользователя %d", money.FromUserID)).
					Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.fromUser, testCase.toUser, testCase.money)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, db)

			err := s.Transfer(testCase.money)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReservation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	type mockBehavior func(
		s *mock_repository.MockControl,
		user *models.User,
		transaction *models.Transaction,
		title string,
		reservBalance int,
		date time.Time)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		transaction   *models.Transaction
		user          *models.User
		service       string
		reservBalance int
		date          time.Time
		wantErr       bool
	}{
		{
			name: "OK",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance-transaction.Amount).Return(nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance+transaction.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					transaction.UserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					transaction.Amount,
					fmt.Sprintf("Заказ №%d, услуга \"%s\"", transaction.OrderID, service)).
					Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "error service not found",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
			},
		},

		{
			name: "error user not found",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error insufficient funds",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 10,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error service",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, errors.New("db error"))
			},
		},

		{
			name: "error getuser",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error getbalance",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service: "Услуга №1",
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalance",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance-transaction.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalance",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance-transaction.Amount).Return(nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance+transaction.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertmoneyreservdetails",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance-transaction.Amount).Return(nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance+transaction.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlog",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance-transaction.Amount).Return(nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance+transaction.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					transaction.UserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					transaction.Amount,
					fmt.Sprintf("Заказ №%d, услуга \"%s\"", transaction.OrderID, service)).
					Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(
				control,
				testCase.user,
				testCase.transaction,
				testCase.service,
				testCase.reservBalance,
				testCase.date)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, db)

			err := s.Reservation(testCase.transaction)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCancelReservation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	type mockBehavior func(
		s *mock_repository.MockControl,
		user *models.User,
		transaction *models.Transaction,
		title string,
		reservBalance int,
		date time.Time,
		rows int64)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		transaction   *models.Transaction
		user          *models.User
		service       string
		reservBalance int
		date          time.Time
		rows          int64
		wantErr       bool
	}{
		{
			name: "OK",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(rows, nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					transaction.UserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					transaction.Amount,
					fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", transaction.OrderID, service)).
					Return(nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance+transaction.Amount).Return(nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance-transaction.Amount).Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "error service not found",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			service: "",
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
			},
		},

		{
			name: "error user not found",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			service: "Услуга №1",
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error reserv not found",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          0,
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(rows, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error service",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, errors.New("db error"))
			},
		},

		{
			name: "error getuser",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			service: "Услуга №1",
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error getbalance",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error delete",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(rows, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlog",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(rows, nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					transaction.UserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					transaction.Amount,
					fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", transaction.OrderID, service)).
					Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalance",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(rows, nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					transaction.UserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					transaction.Amount,
					fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", transaction.OrderID, service)).
					Return(nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance+transaction.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error moneyreservaccounts",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &models.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *models.User,
				transaction *models.Transaction,
				service string,
				reservBalance int,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).
					Return(rows, nil)
				r.EXPECT().InsertLogTx(
					gomock.Any(),
					transaction.UserID,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					transaction.Amount,
					fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", transaction.OrderID, service)).
					Return(nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), transaction.UserID, user.Balance+transaction.Amount).Return(nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance-transaction.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(
				control,
				testCase.user,
				testCase.transaction,
				testCase.service,
				testCase.reservBalance,
				testCase.date,
				testCase.rows)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, db)

			err := s.CancelReservation(testCase.transaction)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfirmation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	type mockBehavior func(
		s *mock_repository.MockControl,
		transaction *models.Transaction,
		reservBalance int,
		date time.Time,
		rows int64)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		transaction   *models.Transaction
		reservBalance int
		rows          int64
		date          time.Time
		wantErr       bool
	}{
		{
			name: "OK",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			mockBehavior: func(
				r *mock_repository.MockControl,
				transaction *models.Transaction,
				reservBalance int,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).Return(rows, nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance-transaction.Amount).Return(nil)
				r.EXPECT().InsertReportTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.Amount,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)).
					Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "error reserv not found",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				transaction *models.Transaction,
				reservBalance int,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).Return(rows, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error getbalance",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				transaction *models.Transaction,
				reservBalance int,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error delete",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				transaction *models.Transaction,
				reservBalance int,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).Return(rows, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				transaction *models.Transaction,
				reservBalance int,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).Return(rows, nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance-transaction.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insert",
			transaction: &models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				transaction *models.Transaction,
				reservBalance int,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, nil)
				r.EXPECT().DeleteMoneyReserveDetailsTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.OrderID,
					transaction.Amount,
					date).Return(rows, nil)
				r.EXPECT().UpdateMoneyReserveAccountsTx(gomock.Any(), transaction.UserID, reservBalance-transaction.Amount).Return(nil)
				r.EXPECT().InsertReportTx(
					gomock.Any(),
					transaction.UserID,
					transaction.ServiceID,
					transaction.Amount,
					time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)).
					Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(
				control,
				testCase.transaction,
				testCase.reservBalance,
				testCase.date,
				testCase.rows)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, db)

			err := s.Confirmation(testCase.transaction)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateReport(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	conf := config.Config{
		Host: "localhost:8081",
	}

	type mockBehavior func(
		s *mock_repository.MockControl,
		requestReport models.RequestReport,
		from time.Time,
		report map[string]int)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		requestReport models.RequestReport
		from          time.Time
		report        map[string]int
		want          string
		wantErr       bool
	}{
		{
			name: "OK",
			requestReport: models.RequestReport{
				Month: 11,
				Year:  2022,
			},
			report: map[string]int{"Услуга №1": 1200},
			from:   time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(
				r *mock_repository.MockControl,
				requestReport models.RequestReport,
				from time.Time,
				report map[string]int) {
				to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
				r.EXPECT().GetReport(from, to).Return(report, nil)
			},
			want: "localhost:8081/file/",
		},

		{
			name: "error",
			requestReport: models.RequestReport{
				Month: 11,
				Year:  2022,
			},
			report:  map[string]int{"Услуга №1": 1200},
			from:    time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				requestReport models.RequestReport,
				from time.Time,
				report map[string]int) {
				to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
				r.EXPECT().GetReport(from, to).Return(report, errors.New("db error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(
				control,
				testCase.requestReport,
				testCase.from,
				testCase.report)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, &conf, db)

			got, err := s.CreateReport(&testCase.requestReport)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got[0:len(got)-14])
				assert.FileExists(t, strings.Replace(got, testCase.want[0:len(testCase.want)-5], "", -1))
			}
		})
	}
}

func TestGetHistory(t *testing.T) {

	type mockBehavior func(s *mock_repository.MockControl, requestHistory *models.RequestHistory)
	db, _, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	testTable := []struct {
		name           string
		mockBehavior   mockBehavior
		requestHistory models.RequestHistory
		want           []models.History
		wantErr        bool
	}{
		{
			name: "OK",
			requestHistory: models.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			want: []models.History{
				{
					Date:        time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
					Amount:      100,
					Description: "Пополнение баланса",
				},
			},
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *models.RequestHistory) {
				r.EXPECT().GetHistory(requestHistory).Return([]models.History{
					{
						Date:        time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
						Amount:      100,
						Description: "Пополнение баланса",
					},
				}, nil)
			},
		},

		{
			name: "error",
			requestHistory: models.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *models.RequestHistory) {
				r.EXPECT().GetHistory(requestHistory).Return(nil, errors.New("db error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.requestHistory)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository, nil, db)

			got, err := s.GetHistory(&testCase.requestHistory)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}
