package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
	"userbalance/internal/config"
	"userbalance/internal/repository"
	mock_repository "userbalance/internal/repository/mocks"
	"userbalance/pkg/api"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetBalance(t *testing.T) {

	type mockBehavior func(s *mock_repository.MockControl, userId int32)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		userId       int32
		want         *api.User
		wantErr      bool
	}{
		{
			name:   "OK",
			userId: 1,
			mockBehavior: func(r *mock_repository.MockControl, userId int32) {
				r.EXPECT().GetUser(userId).Return(
					&api.User{
						Id:      1,
						Balance: 100}, nil)
			},
			want: &api.User{
				Id:      1,
				Balance: 100,
			},
		},

		{
			name:    "error user not found",
			userId:  0,
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, userId int32) {
				r.EXPECT().GetUser(userId).Return(
					nil, errors.New("пользователь не найден"))
			},
		},

		{
			name:    "error database",
			userId:  0,
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, userId int32) {
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

	type mockBehavior func(s *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User)

	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		replenishment *api.Replenishment
		user          *api.User
		date          time.Time
		wantErr       bool
	}{
		{
			name: "OK user exists",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			date: time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			user: &api.User{
				Id:      1,
				Balance: 200,
			},
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(user, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), replenishment.UserID, user.Balance+replenishment.Amount).Return(nil)
				r.EXPECT().InsertLogTx(gomock.Any(), replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(nil)
				mock.ExpectCommit()
			},
		},

		{
			name: "OK user not found",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			date: time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
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
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalance",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 200,
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(user, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), replenishment.UserID, user.Balance+replenishment.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlog",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 200,
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(user, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), replenishment.UserID, user.Balance+replenishment.Amount).Return(nil)
				r.EXPECT().InsertLogTx(gomock.Any(), replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertuser",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, nil)
				r.EXPECT().InsertUserTx(gomock.Any(), replenishment.UserID, replenishment.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertmoneyreservacounts",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), replenishment.UserID).Return(nil, nil)
				r.EXPECT().InsertUserTx(gomock.Any(), replenishment.UserID, replenishment.Amount).Return(nil)
				r.EXPECT().InsertMoneyReserveAccountsTx(gomock.Any(), replenishment.UserID).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlog",
			replenishment: &api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(r *mock_repository.MockControl, replenishment *api.Replenishment, user *api.User) {
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

	type mockBehavior func(s *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		money        *api.Money
		fromUser     *api.User
		toUser       *api.User
		wantErr      bool
	}{
		{
			name: "OK",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
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
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error touser not found",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error insufficient funds",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     2000,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error getfromuser",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     2000,
				Date:       "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error gettouser",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     2000,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error updatebalancefromuser",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.FromUserID).Return(fromUser, nil)
				r.EXPECT().GetUserForUpdate(gomock.Any(), money.ToUserID).Return(toUser, nil)
				r.EXPECT().UpdateBalanceTx(gomock.Any(), money.FromUserID, fromUser.Balance-money.Amount).Return(errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insertlogfromuser",
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
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
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
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
			money: &api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			fromUser: &api.User{
				Id:      1,
				Balance: 1000,
			},
			toUser: &api.User{
				Id:      2,
				Balance: 500,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, fromUser *api.User, toUser *api.User, money *api.Money) {
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
		user *api.User,
		transaction *api.Transaction,
		title string,
		reservBalance int32,
		date time.Time)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		transaction   *api.Transaction
		user          *api.User
		service       string
		reservBalance int32
		date          time.Time
		wantErr       bool
	}{
		{
			name: "OK",
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
			},
		},

		{
			name: "error user not found",
			transaction: &api.Transaction{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(nil, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error insufficient funds",
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 10,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(user, nil)
				mock.ExpectRollback()
			},
		},

		{
			name: "error service",
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, errors.New("db error"))
			},
		},

		{
			name: "error getuser",
			transaction: &api.Transaction{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
				mock.ExpectBegin()
				r.EXPECT().GetUserForUpdate(gomock.Any(), transaction.UserID).Return(nil, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error getbalance",
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service: "Услуга №1",
			date:    time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			service:       "Услуга №1",
			reservBalance: 1000,
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
		user *api.User,
		transaction *api.Transaction,
		title string,
		reservBalance int32,
		date time.Time,
		rows int64)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		transaction   *api.Transaction
		user          *api.User
		service       string
		reservBalance int32
		date          time.Time
		rows          int64
		wantErr       bool
	}{
		{
			name: "OK",
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			rows:          1,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, nil)
			},
		},

		{
			name: "error user not found",
			transaction: &api.Transaction{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
				date time.Time,
				rows int64) {
				r.EXPECT().GetService(transaction.ServiceID).Return(service, errors.New("db error"))
			},
		},

		{
			name: "error getuser",
			transaction: &api.Transaction{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
				Id:      1,
				Balance: 1000,
			},
			reservBalance: 1000,
			service:       "Услуга №1",
			date:          time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC),
			wantErr:       true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
			transaction: &api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &api.User{
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
				user *api.User,
				transaction *api.Transaction,
				service string,
				reservBalance int32,
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
		transaction *api.Transaction,
		reservBalance int32,
		date time.Time,
		rows int64)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		transaction   *api.Transaction
		reservBalance int32
		rows          int64
		date          time.Time
		wantErr       bool
	}{
		{
			name: "OK",
			transaction: &api.Transaction{
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
				transaction *api.Transaction,
				reservBalance int32,
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
			transaction: &api.Transaction{
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
				transaction *api.Transaction,
				reservBalance int32,
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
			transaction: &api.Transaction{
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
				transaction *api.Transaction,
				reservBalance int32,
				date time.Time,
				rows int64) {
				mock.ExpectBegin()
				r.EXPECT().GetBalanceReserveAccountsTx(gomock.Any(), transaction.UserID).Return(reservBalance, errors.New("db error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error delete",
			transaction: &api.Transaction{
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
				transaction *api.Transaction,
				reservBalance int32,
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
			transaction: &api.Transaction{
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
				transaction *api.Transaction,
				reservBalance int32,
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
			transaction: &api.Transaction{
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
				transaction *api.Transaction,
				reservBalance int32,
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
		requestReport api.RequestReport,
		from time.Time,
		report map[string]int32)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		requestReport api.RequestReport
		from          time.Time
		report        map[string]int32
		want          string
		wantErr       bool
	}{
		{
			name: "OK",
			requestReport: api.RequestReport{
				Month: 11,
				Year:  2022,
			},
			report: map[string]int32{"Услуга №1": 1200},
			from:   time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			mockBehavior: func(
				r *mock_repository.MockControl,
				requestReport api.RequestReport,
				from time.Time,
				report map[string]int32) {
				to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
				r.EXPECT().GetReport(from, to).Return(report, nil)
			},
			want: "localhost:8081/file/",
		},

		{
			name: "error",
			requestReport: api.RequestReport{
				Month: 11,
				Year:  2022,
			},
			report:  map[string]int32{"Услуга №1": 1200},
			from:    time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			wantErr: true,
			mockBehavior: func(
				r *mock_repository.MockControl,
				requestReport api.RequestReport,
				from time.Time,
				report map[string]int32) {
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

	type mockBehavior func(s *mock_repository.MockControl, requestHistory *api.RequestHistory)
	db, _, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	testTable := []struct {
		name           string
		mockBehavior   mockBehavior
		requestHistory api.RequestHistory
		want           []*api.History
		wantErr        bool
	}{
		{
			name: "OK",
			requestHistory: api.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			want: []*api.History{
				{
					Date:        &timestamppb.Timestamp{Seconds: time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local).Unix()},
					Amount:      100,
					Description: "Пополнение баланса",
				},
			},
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *api.RequestHistory) {
				r.EXPECT().GetHistory(requestHistory).Return([]*api.History{
					{
						Date:        &timestamppb.Timestamp{Seconds: time.Date(2022, 10, 01, 0, 0, 0, 0, time.Local).Unix()},
						Amount:      100,
						Description: "Пополнение баланса",
					},
				}, nil)
			},
		},

		{
			name: "error",
			requestHistory: api.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *api.RequestHistory) {
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
