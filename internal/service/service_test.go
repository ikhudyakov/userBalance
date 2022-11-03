package service

import (
	"errors"
	"strings"
	"testing"
	"userbalance"
	"userbalance/internal/repository"
	mock_repository "userbalance/internal/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {

	type mockBehavior func(s *mock_repository.MockControl, userId int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		userId       int
		want         userbalance.User
		wantErr      bool
	}{
		{
			name:   "OK",
			userId: 1,
			mockBehavior: func(r *mock_repository.MockControl, userId int) {
				r.EXPECT().GetBalance(userId).Return(
					userbalance.User{
						Id:      1,
						Balance: 100}, nil)
			},
			want: userbalance.User{
				Id:      1,
				Balance: 100,
			},
		},

		{
			name:    "error user not found",
			userId:  0,
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, userId int) {
				r.EXPECT().GetBalance(userId).Return(
					userbalance.User{}, nil)
			},
			want: userbalance.User{},
		},

		{
			name:    "error database",
			userId:  0,
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, userId int) {
				r.EXPECT().GetBalance(userId).Return(
					userbalance.User{}, errors.New("error database"))
			},
			want: userbalance.User{},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.userId)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

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

	type mockBehavior func(s *mock_repository.MockControl, transaction *userbalance.Transaction)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		transaction  *userbalance.Transaction
		wantErr      bool
	}{
		{
			name: "OK",
			transaction: &userbalance.Transaction{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {
				r.EXPECT().ReplenishmentBalance(transaction.UserID, transaction.Amount, transaction.Date).Return(nil)
			},
		},

		{
			name: "error userid <= 0",
			transaction: &userbalance.Transaction{
				UserID: 0,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr:      true,
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {},
		},

		{
			name: "error amount <= 0",
			transaction: &userbalance.Transaction{
				UserID: 1,
				Amount: -10,
				Date:   "2022-10-01",
			},
			wantErr:      true,
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {},
		},

		{
			name: "error database",
			transaction: &userbalance.Transaction{
				UserID: 1,
				Amount: 100,
				Date:   "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {
				r.EXPECT().ReplenishmentBalance(transaction.UserID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.transaction)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

			err := s.ReplenishmentBalance(testCase.transaction)

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransfer(t *testing.T) {

	type mockBehavior func(s *mock_repository.MockControl, user *userbalance.User, money *userbalance.Money)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		money        *userbalance.Money
		user         *userbalance.User
		wantErr      bool
	}{
		{
			name: "OK",
			money: &userbalance.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 1000,
			},
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, money *userbalance.Money) {
				r.EXPECT().GetBalance(money.FromUserID).Return(*user, nil)
				r.EXPECT().Transfer(money.FromUserID, money.ToUserID, money.Amount, money.Date).Return(nil)
			},
		},

		{
			name: "error amount <= 0",
			money: &userbalance.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     -10,
				Date:       "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr:      true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, money *userbalance.Money) {},
		},

		{
			name: "error database",
			money: &userbalance.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, money *userbalance.Money) {
				r.EXPECT().GetBalance(money.FromUserID).Return(*user, nil)
				r.EXPECT().Transfer(money.FromUserID, money.ToUserID, money.Amount, money.Date).Return(errors.New("error database"))
			},
		},

		{
			name: "error insufficient funds",
			money: &userbalance.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 10,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, money *userbalance.Money) {
				r.EXPECT().GetBalance(money.FromUserID).Return(*user, nil)
			},
		},

		{
			name: "error fromuserid == touserid",
			money: &userbalance.Money{
				FromUserID: 1,
				ToUserID:   1,
				Amount:     100,
				Date:       "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 10,
			},
			wantErr:      true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, money *userbalance.Money) {},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.user, testCase.money)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

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

	type mockBehavior func(s *mock_repository.MockControl, user *userbalance.User, transaction *userbalance.Transaction)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		transaction  *userbalance.Transaction
		user         *userbalance.User
		wantErr      bool
	}{
		{
			name: "OK",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 1000,
			},
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, transaction *userbalance.Transaction) {
				r.EXPECT().GetBalance(transaction.UserID).Return(*user, nil)
				r.EXPECT().Reservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(nil)
			},
		},

		{
			name: "error orderId <= 0",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   0,
				Date:      "2022-10-01",
			},
			user:         &userbalance.User{},
			wantErr:      true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, transaction *userbalance.Transaction) {},
		},

		{
			name: "error amount <= 0",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    -10,
				ServiceID: 1,
				OrderID:   1,
				Date:      "2022-10-01",
			},
			user:         &userbalance.User{},
			wantErr:      true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, transaction *userbalance.Transaction) {},
		},

		{
			name: "error insufficient funds",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 10,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, transaction *userbalance.Transaction) {
				r.EXPECT().GetBalance(transaction.UserID).Return(*user, nil)
			},
		},

		{
			name: "error database",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			user: &userbalance.User{
				Id:      1,
				Balance: 1000,
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, user *userbalance.User, transaction *userbalance.Transaction) {
				r.EXPECT().GetBalance(transaction.UserID).Return(*user, nil)
				r.EXPECT().Reservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.user, testCase.transaction)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

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

	type mockBehavior func(s *mock_repository.MockControl, transaction *userbalance.Transaction)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		transaction  *userbalance.Transaction
		wantErr      bool
	}{
		{
			name: "OK",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {
				r.EXPECT().CancelReservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(nil)
			},
		},

		{
			name: "error database",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {
				r.EXPECT().CancelReservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.transaction)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

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

	type mockBehavior func(s *mock_repository.MockControl, transaction *userbalance.Transaction)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		transaction  *userbalance.Transaction
		wantErr      bool
	}{
		{
			name: "OK",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {
				r.EXPECT().Confirmation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(nil)
			},
		},

		{
			name: "error database",
			transaction: &userbalance.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   10,
				Date:      "2022-10-01",
			},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, transaction *userbalance.Transaction) {
				r.EXPECT().Confirmation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.transaction)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

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

	type mockBehavior func(s *mock_repository.MockControl, requestReport userbalance.RequestReport)

	testTable := []struct {
		name          string
		mockBehavior  mockBehavior
		requestReport userbalance.RequestReport
		report        userbalance.Report
		want          string
		wantErr       bool
	}{
		{
			name: "OK",
			requestReport: userbalance.RequestReport{
				FromDate: "2022-10-01",
				ToDate:   "2022-10-31",
			},
			report: userbalance.Report{
				Title:  "Услуга №1",
				Amount: 1200,
			},
			mockBehavior: func(r *mock_repository.MockControl, requestReport userbalance.RequestReport) {
				r.EXPECT().CreateReport(requestReport.FromDate, requestReport.ToDate).Return(map[string]int{"Услуга №1": 1200}, nil)
			},
			want: "localhost:8081/file/",
		},

		{
			name: "error database",
			requestReport: userbalance.RequestReport{
				FromDate: "2022-10-01",
				ToDate:   "2022-10-31",
			},
			report: userbalance.Report{},
			mockBehavior: func(r *mock_repository.MockControl, requestReport userbalance.RequestReport) {
				r.EXPECT().CreateReport(requestReport.FromDate, requestReport.ToDate).Return(map[string]int{}, errors.New("error database"))
			},
			want:    "localhost:8081/file/",
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_repository.NewMockControl(c)
			testCase.mockBehavior(control, testCase.requestReport)

			repository := &repository.Repository{Control: control}
			s := NewControlService(repository)

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

	type mockBehavior func(s *mock_repository.MockControl, requestHistory *userbalance.RequestHistory)

	testTable := []struct {
		name           string
		mockBehavior   mockBehavior
		requestHistory userbalance.RequestHistory
		want           []userbalance.History
		wantErr        bool
	}{
		{
			name: "OK",
			requestHistory: userbalance.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			want: []userbalance.History{
				{
					Date:        "01/10/2022",
					Amount:      100,
					Description: "Пополнение баланса",
				},
			},
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *userbalance.RequestHistory) {
				r.EXPECT().GetHistory(requestHistory.UserID).Return([]userbalance.History{
					{
						Date:        "01/10/2022",
						Amount:      100,
						Description: "Пополнение баланса",
					},
				}, nil)
			},
		},

		{
			name: "error history not found",
			requestHistory: userbalance.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			want:    []userbalance.History{},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *userbalance.RequestHistory) {
				r.EXPECT().GetHistory(requestHistory.UserID).Return([]userbalance.History{}, nil)
			},
		},

		{
			name: "error database",
			requestHistory: userbalance.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			want:    []userbalance.History{},
			wantErr: true,
			mockBehavior: func(r *mock_repository.MockControl, requestHistory *userbalance.RequestHistory) {
				r.EXPECT().GetHistory(requestHistory.UserID).Return([]userbalance.History{}, errors.New("error database"))
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
			s := NewControlService(repository)

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
