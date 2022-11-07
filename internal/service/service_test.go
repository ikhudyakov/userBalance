package service

import (
	"errors"
	"log"
	"testing"
	"time"
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

	tx, err := db.Begin()
	log.Println(err)

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
				r.EXPECT().GetUserForUpdate(tx, replenishment.UserID).Return(user, nil)
				if user != nil {
					r.EXPECT().UpdateBalanceTx(tx, replenishment.UserID, user.Balance+replenishment.Amount).Return(nil)
					r.EXPECT().InsertLogTx(tx, replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(nil)
				} else {
					r.EXPECT().InsertUserTx(tx, replenishment.UserID, replenishment.Amount).Return(nil)
					r.EXPECT().InsertMoneyReserveAccountsTx(tx, replenishment.UserID).Return(nil)
					r.EXPECT().InsertLogTx(tx, replenishment.UserID, time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC), replenishment.Amount, "Пополнение баланса").Return(nil)
				}
			},
		},

		// {
		// 	name: "error userid <= 0",
		// 	transaction: &models.Transaction{
		// 		UserID: 0,
		// 		Amount: 100,
		// 		Date:   "2022-10-01",
		// 	},
		// 	wantErr:      true,
		// 	mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {},
		// },

		// {
		// 	name: "error amount <= 0",
		// 	transaction: &models.Transaction{
		// 		UserID: 1,
		// 		Amount: -10,
		// 		Date:   "2022-10-01",
		// 	},
		// 	wantErr:      true,
		// 	mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {},
		// },

		// {
		// 	name: "error database",
		// 	transaction: &models.Transaction{
		// 		UserID: 1,
		// 		Amount: 100,
		// 		Date:   "2022-10-01",
		// 	},
		// 	wantErr: true,
		// 	mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {
		// 		r.EXPECT().ReplenishmentBalance(transaction.UserID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
		// 	},
		// },
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

			mock.ExpectCommit()

			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// func TestTransfer(t *testing.T) {

// 	type mockBehavior func(s *mock_repository.MockControl, user *models.User, money *models.Money)

// 	testTable := []struct {
// 		name         string
// 		mockBehavior mockBehavior
// 		money        *models.Money
// 		user         *models.User
// 		wantErr      bool
// 	}{
// 		{
// 			name: "OK",
// 			money: &models.Money{
// 				FromUserID: 1,
// 				ToUserID:   2,
// 				Amount:     100,
// 				Date:       "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 1000,
// 			},
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, money *models.Money) {
// 				r.EXPECT().GetBalance(money.FromUserID).Return(*user, nil)
// 				r.EXPECT().Transfer(money.FromUserID, money.ToUserID, money.Amount, money.Date).Return(nil)
// 			},
// 		},

// 		{
// 			name: "error amount <= 0",
// 			money: &models.Money{
// 				FromUserID: 1,
// 				ToUserID:   2,
// 				Amount:     -10,
// 				Date:       "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 1000,
// 			},
// 			wantErr:      true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, money *models.Money) {},
// 		},

// 		{
// 			name: "error database",
// 			money: &models.Money{
// 				FromUserID: 1,
// 				ToUserID:   2,
// 				Amount:     100,
// 				Date:       "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 1000,
// 			},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, money *models.Money) {
// 				r.EXPECT().GetBalance(money.FromUserID).Return(*user, nil)
// 				r.EXPECT().Transfer(money.FromUserID, money.ToUserID, money.Amount, money.Date).Return(errors.New("error database"))
// 			},
// 		},

// 		{
// 			name: "error insufficient funds",
// 			money: &models.Money{
// 				FromUserID: 1,
// 				ToUserID:   2,
// 				Amount:     100,
// 				Date:       "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 10,
// 			},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, money *models.Money) {
// 				r.EXPECT().GetBalance(money.FromUserID).Return(*user, nil)
// 			},
// 		},

// 		{
// 			name: "error fromuserid == touserid",
// 			money: &models.Money{
// 				FromUserID: 1,
// 				ToUserID:   1,
// 				Amount:     100,
// 				Date:       "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 10,
// 			},
// 			wantErr:      true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, money *models.Money) {},
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_repository.NewMockControl(c)
// 			testCase.mockBehavior(control, testCase.user, testCase.money)

// 			repository := &repository.Repository{Control: control}
// 			s := NewControlService(repository)

// 			err := s.Transfer(testCase.money)

// 			if testCase.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestReservation(t *testing.T) {

// 	type mockBehavior func(s *mock_repository.MockControl, user *models.User, transaction *models.Transaction)

// 	testTable := []struct {
// 		name         string
// 		mockBehavior mockBehavior
// 		transaction  *models.Transaction
// 		user         *models.User
// 		wantErr      bool
// 	}{
// 		{
// 			name: "OK",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 1000,
// 			},
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, transaction *models.Transaction) {
// 				r.EXPECT().GetBalance(transaction.UserID).Return(*user, nil)
// 				r.EXPECT().Reservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(nil)
// 			},
// 		},

// 		{
// 			name: "error orderId <= 0",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   0,
// 				Date:      "2022-10-01",
// 			},
// 			user:         &models.User{},
// 			wantErr:      true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, transaction *models.Transaction) {},
// 		},

// 		{
// 			name: "error amount <= 0",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    -10,
// 				ServiceID: 1,
// 				OrderID:   1,
// 				Date:      "2022-10-01",
// 			},
// 			user:         &models.User{},
// 			wantErr:      true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, transaction *models.Transaction) {},
// 		},

// 		{
// 			name: "error insufficient funds",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 10,
// 			},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, transaction *models.Transaction) {
// 				r.EXPECT().GetBalance(transaction.UserID).Return(*user, nil)
// 			},
// 		},

// 		{
// 			name: "error database",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			user: &models.User{
// 				Id:      1,
// 				Balance: 1000,
// 			},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, user *models.User, transaction *models.Transaction) {
// 				r.EXPECT().GetBalance(transaction.UserID).Return(*user, nil)
// 				r.EXPECT().Reservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
// 			},
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_repository.NewMockControl(c)
// 			testCase.mockBehavior(control, testCase.user, testCase.transaction)

// 			repository := &repository.Repository{Control: control}
// 			s := NewControlService(repository)

// 			err := s.Reservation(testCase.transaction)

// 			if testCase.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestCancelReservation(t *testing.T) {

// 	type mockBehavior func(s *mock_repository.MockControl, transaction *models.Transaction)

// 	testTable := []struct {
// 		name         string
// 		mockBehavior mockBehavior
// 		transaction  *models.Transaction
// 		wantErr      bool
// 	}{
// 		{
// 			name: "OK",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {
// 				r.EXPECT().CancelReservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(nil)
// 			},
// 		},

// 		{
// 			name: "error database",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {
// 				r.EXPECT().CancelReservation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
// 			},
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_repository.NewMockControl(c)
// 			testCase.mockBehavior(control, testCase.transaction)

// 			repository := &repository.Repository{Control: control}
// 			s := NewControlService(repository)

// 			err := s.CancelReservation(testCase.transaction)

// 			if testCase.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestConfirmation(t *testing.T) {

// 	type mockBehavior func(s *mock_repository.MockControl, transaction *models.Transaction)

// 	testTable := []struct {
// 		name         string
// 		mockBehavior mockBehavior
// 		transaction  *models.Transaction
// 		wantErr      bool
// 	}{
// 		{
// 			name: "OK",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {
// 				r.EXPECT().Confirmation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(nil)
// 			},
// 		},

// 		{
// 			name: "error database",
// 			transaction: &models.Transaction{
// 				UserID:    1,
// 				Amount:    100,
// 				ServiceID: 1,
// 				OrderID:   10,
// 				Date:      "2022-10-01",
// 			},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, transaction *models.Transaction) {
// 				r.EXPECT().Confirmation(transaction.UserID, transaction.ServiceID, transaction.OrderID, transaction.Amount, transaction.Date).Return(errors.New("error database"))
// 			},
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_repository.NewMockControl(c)
// 			testCase.mockBehavior(control, testCase.transaction)

// 			repository := &repository.Repository{Control: control}
// 			s := NewControlService(repository)

// 			err := s.Confirmation(testCase.transaction)

// 			if testCase.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestCreateReport(t *testing.T) {

// 	type mockBehavior func(s *mock_repository.MockControl, requestReport models.RequestReport)

// 	testTable := []struct {
// 		name          string
// 		mockBehavior  mockBehavior
// 		requestReport models.RequestReport
// 		report        models.Report
// 		want          string
// 		wantErr       bool
// 	}{
// 		{
// 			name: "OK",
// 			requestReport: models.RequestReport{
// 				FromDate: "2022-10-01",
// 				ToDate:   "2022-10-31",
// 			},
// 			report: models.Report{
// 				Title:  "Услуга №1",
// 				Amount: 1200,
// 			},
// 			mockBehavior: func(r *mock_repository.MockControl, requestReport models.RequestReport) {
// 				r.EXPECT().CreateReport(requestReport.FromDate, requestReport.ToDate).Return(map[string]int{"Услуга №1": 1200}, nil)
// 			},
// 			want: "localhost:8081/file/",
// 		},

// 		{
// 			name: "error database",
// 			requestReport: models.RequestReport{
// 				FromDate: "2022-10-01",
// 				ToDate:   "2022-10-31",
// 			},
// 			report: models.Report{},
// 			mockBehavior: func(r *mock_repository.MockControl, requestReport models.RequestReport) {
// 				r.EXPECT().CreateReport(requestReport.FromDate, requestReport.ToDate).Return(map[string]int{}, errors.New("error database"))
// 			},
// 			want:    "localhost:8081/file/",
// 			wantErr: true,
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_repository.NewMockControl(c)
// 			testCase.mockBehavior(control, testCase.requestReport)

// 			repository := &repository.Repository{Control: control}
// 			s := NewControlService(repository)

// 			got, err := s.CreateReport(&testCase.requestReport)

// 			if testCase.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, testCase.want, got[0:len(got)-14])
// 				assert.FileExists(t, strings.Replace(got, testCase.want[0:len(testCase.want)-5], "", -1))
// 			}
// 		})
// 	}
// }

// func TestGetHistory(t *testing.T) {

// 	type mockBehavior func(s *mock_repository.MockControl, requestHistory *models.RequestHistory)

// 	testTable := []struct {
// 		name           string
// 		mockBehavior   mockBehavior
// 		requestHistory models.RequestHistory
// 		want           []models.History
// 		wantErr        bool
// 	}{
// 		{
// 			name: "OK",
// 			requestHistory: models.RequestHistory{
// 				UserID:    1,
// 				SortField: "amount",
// 				Direction: "desc",
// 			},
// 			want: []models.History{
// 				{
// 					Date:        "01/10/2022",
// 					Amount:      100,
// 					Description: "Пополнение баланса",
// 				},
// 			},
// 			mockBehavior: func(r *mock_repository.MockControl, requestHistory *models.RequestHistory) {
// 				r.EXPECT().GetHistory(requestHistory.UserID).Return([]models.History{
// 					{
// 						Date:        "01/10/2022",
// 						Amount:      100,
// 						Description: "Пополнение баланса",
// 					},
// 				}, nil)
// 			},
// 		},

// 		{
// 			name: "error history not found",
// 			requestHistory: models.RequestHistory{
// 				UserID:    1,
// 				SortField: "amount",
// 				Direction: "desc",
// 			},
// 			want:    []models.History{},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, requestHistory *models.RequestHistory) {
// 				r.EXPECT().GetHistory(requestHistory.UserID).Return([]models.History{}, nil)
// 			},
// 		},

// 		{
// 			name: "error database",
// 			requestHistory: models.RequestHistory{
// 				UserID:    1,
// 				SortField: "amount",
// 				Direction: "desc",
// 			},
// 			want:    []models.History{},
// 			wantErr: true,
// 			mockBehavior: func(r *mock_repository.MockControl, requestHistory *models.RequestHistory) {
// 				r.EXPECT().GetHistory(requestHistory.UserID).Return([]models.History{}, errors.New("error database"))
// 			},
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_repository.NewMockControl(c)
// 			testCase.mockBehavior(control, &testCase.requestHistory)

// 			repository := &repository.Repository{Control: control}
// 			s := NewControlService(repository)

// 			got, err := s.GetHistory(&testCase.requestHistory)

// 			if testCase.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, testCase.want, got)
// 			}
// 		})
// 	}
// }
