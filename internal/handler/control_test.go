package handler

import (
	"context"
	"errors"
	"testing"
	"time"
	"userbalance/internal/service"
	mock_service "userbalance/internal/service/mocks"
	"userbalance/pkg/api"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_GetBalance(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, user *api.User)

	testTable := []struct {
		name         string
		ctx          context.Context
		inputUser    api.User
		mockBehavior mockBehavior
		want         *api.User
		wantErr      bool
	}{
		{
			name: "OK",
			ctx:  context.Background(),
			inputUser: api.User{
				Id: 1,
			},
			mockBehavior: func(s *mock_service.MockControl, user *api.User) {
				s.EXPECT().GetBalance(user.Id).Return(
					&api.User{
						Id:      1,
						Balance: 100}, nil)
			},
			want:    &api.User{Id: 1, Balance: 100},
			wantErr: false,
		},

		{
			name: "error wrong userid",
			ctx:  context.Background(),
			inputUser: api.User{
				Id: 10,
			},
			mockBehavior: func(s *mock_service.MockControl, user *api.User) {
				s.EXPECT().GetBalance(user.Id).Return(
					nil, errors.New("wrong userid"))
			},
			wantErr: true,
		},

		{
			name:         "error empty field",
			ctx:          context.Background(),
			inputUser:    api.User{},
			mockBehavior: func(s *mock_service.MockControl, user *api.User) {},
			wantErr:      true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputUser)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.GetBalance(
				testCase.ctx, &testCase.inputUser)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestHandler_ReplenishmentBalance(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, replenishment *api.Replenishment)

	testTable := []struct {
		name               string
		ctx                context.Context
		inputReplenishment api.Replenishment
		mockBehavior       mockBehavior
		want               *api.Response
		wantErr            bool
	}{
		{
			name: "OK",
			ctx:  context.Background(),
			inputReplenishment: api.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-11-01",
			},
			mockBehavior: func(s *mock_service.MockControl, replenishment *api.Replenishment) {
				s.EXPECT().ReplenishmentBalance(replenishment).Return(nil)
			},
			want:    &api.Response{Message: "баланс пополнен"},
			wantErr: false,
		},

		{
			name: "error userId <= 0",
			ctx:  context.Background(),
			inputReplenishment: api.Replenishment{
				UserID: -1,
				Amount: 100,
				Date:   "2022-11-01",
			},
			mockBehavior: func(s *mock_service.MockControl, replenishment *api.Replenishment) {
			},
			wantErr: true,
		},

		{
			name: "error amount <= 0",
			ctx:  context.Background(),
			inputReplenishment: api.Replenishment{
				UserID: 1,
				Amount: -100,
				Date:   "2022-11-01",
			},
			mockBehavior: func(s *mock_service.MockControl, replenishment *api.Replenishment) {
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputReplenishment)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.ReplenishmentBalance(
				testCase.ctx, &testCase.inputReplenishment)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestHandler_Transfer(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, money *api.Money)

	testTable := []struct {
		name         string
		ctx          context.Context
		inputMoney   api.Money
		mockBehavior mockBehavior
		want         *api.Response
		wantErr      bool
	}{
		{
			name: "OK",
			ctx:  context.Background(),
			inputMoney: api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-08-01",
			},
			mockBehavior: func(s *mock_service.MockControl, money *api.Money) {
				s.EXPECT().Transfer(money).Return(nil)
			},
			want:    &api.Response{Message: "перевод стредств выполнен"},
			wantErr: false,
		},

		{
			name: "error fromUserId <=0",
			ctx:  context.Background(),
			inputMoney: api.Money{
				FromUserID: -1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-08-01",
			},
			mockBehavior: func(s *mock_service.MockControl, money *api.Money) {},
			wantErr:      true,
		},

		{
			name: "error toUserId <=0",
			ctx:  context.Background(),
			inputMoney: api.Money{
				FromUserID: 1,
				ToUserID:   -2,
				Amount:     100,
				Date:       "2022-08-01",
			},
			mockBehavior: func(s *mock_service.MockControl, money *api.Money) {},
			wantErr:      true,
		},

		{
			name: "error toUserId == fromUserId",
			ctx:  context.Background(),
			inputMoney: api.Money{
				FromUserID: 1,
				ToUserID:   1,
				Amount:     100,
				Date:       "2022-08-01",
			},
			mockBehavior: func(s *mock_service.MockControl, money *api.Money) {},
			wantErr:      true,
		},

		{
			name: "error amount <=0",
			ctx:  context.Background(),
			inputMoney: api.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     -100,
				Date:       "2022-08-01",
			},
			mockBehavior: func(s *mock_service.MockControl, money *api.Money) {},
			wantErr:      true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputMoney)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.Transfer(
				testCase.ctx, &testCase.inputMoney)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestHandler_GetHistory(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, requestHistory *api.RequestHistory)

	testTable := []struct {
		name                string
		ctx                 context.Context
		inputRequestHistory api.RequestHistory
		mockBehavior        mockBehavior
		want                *api.Histories
		wantErr             bool
	}{
		{
			name: "OK",
			inputRequestHistory: api.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			mockBehavior: func(s *mock_service.MockControl, requestHistory *api.RequestHistory) {
				s.EXPECT().GetHistory(requestHistory).Return([]*api.History{{
					Date:        &timestamppb.Timestamp{Seconds: (time.Date(2022, 11, 01, 0, 0, 0, 0, time.UTC).Unix())},
					Amount:      500,
					Description: "Пополнение баланса"}}, nil)
			},
			want:    &api.Histories{Entity: []*api.History{{Date: &timestamppb.Timestamp{Seconds: 1667260800}, Amount: 500, Description: "Пополнение баланса"}}},
			wantErr: false,
		},

		{
			name: "error userId <= 0",
			inputRequestHistory: api.RequestHistory{
				UserID: 0,
			},
			mockBehavior: func(s *mock_service.MockControl, requestHistory *api.RequestHistory) {
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputRequestHistory)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.GetHistory(
				testCase.ctx, &testCase.inputRequestHistory)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestHandler_Reservation(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, transaction *api.Transaction)

	testTable := []struct {
		name             string
		ctx              context.Context
		inputTransaction api.Transaction
		mockBehavior     mockBehavior
		want             *api.Response
		wantErr          bool
	}{
		{
			name: "OK",
			ctx:  context.Background(),
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {
				s.EXPECT().Reservation(transaction).Return(nil)
			},
			want:    &api.Response{Message: "резервирование средств прошло успешно"},
			wantErr: false,
		},

		{
			name: "error user <= 0",
			ctx:  context.Background(),
			inputTransaction: api.Transaction{
				UserID:    0,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error amount <= 0",
			ctx:  context.Background(),
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    0,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error serviceid <= 0",
			ctx:  context.Background(),
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: -1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error orderid <= 0",
			ctx:  context.Background(),
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   -1,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputTransaction)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.Reservation(
				testCase.ctx, &testCase.inputTransaction)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestHandler_Confirmation(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, transaction *api.Transaction)

	testTable := []struct {
		name             string
		ctx              context.Context
		inputTransaction api.Transaction
		mockBehavior     mockBehavior
		want             *api.Response
		wantErr          bool
	}{
		{
			name: "OK",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {
				s.EXPECT().Confirmation(transaction).Return(nil)
			},
			want:    &api.Response{Message: "средства из резерва были списаны успешно"},
			wantErr: false,
		},

		{
			name: "error user <= 0",
			inputTransaction: api.Transaction{
				UserID:    0,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error amount <= 0",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    0,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error serviceid <= 0",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: -1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error orderid <= 0",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   -1,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputTransaction)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.Confirmation(
				testCase.ctx, &testCase.inputTransaction)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestHandler_CancelReservation(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, transaction *api.Transaction)

	testTable := []struct {
		name             string
		ctx              context.Context
		inputTransaction api.Transaction
		mockBehavior     mockBehavior
		want             *api.Response
		wantErr          bool
	}{
		{
			name: "OK",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {
				s.EXPECT().CancelReservation(transaction).Return(nil)
			},
			want:    &api.Response{Message: "разрезервирование средств прошло успешно"},
			wantErr: false,
		},

		{
			name: "error user <= 0",
			inputTransaction: api.Transaction{
				UserID:    0,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error amount <= 0",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    0,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error serviceid <= 0",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: -1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},

		{
			name: "error orderid <= 0",
			inputTransaction: api.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   -12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction *api.Transaction) {},
			wantErr:      true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, &testCase.inputTransaction)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			got, err := h.CancelReservation(
				testCase.ctx, &testCase.inputTransaction)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

// func TestHandler_createReport(t *testing.T) {

// 	type mockBehavior func(s *mock_service.MockControl, requestReport models.RequestReport)

// 	testTable := []struct {
// 		name                string
// 		inputBody           string
// 		inputRequestReport  models.RequestReport
// 		mockBehavior        mockBehavior
// 		expectedStatusCode  int
// 		expectedRequestBody string
// 	}{
// 		{
// 			name:      "OK",
// 			inputBody: `{"month":11,"year":2022}`,
// 			inputRequestReport: models.RequestReport{
// 				Month: 11,
// 				Year:  2022,
// 			},
// 			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
// 				s.EXPECT().CreateReport(&requestReport).Return("localhost:8081/file/report.cvs", nil)
// 			},
// 			expectedStatusCode:  http.StatusOK,
// 			expectedRequestBody: `{"message":"localhost:8081/file/report.cvs"}`,
// 		},

// 		{
// 			name:      "error wrong month",
// 			inputBody: `{"month":22,"year":2022}`,
// 			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
// 			},
// 			expectedStatusCode:  http.StatusBadRequest,
// 			expectedRequestBody: `{"message":"month: месяца не может быть \u003e 12."}`,
// 		},

// 		{
// 			name:      "error wrong year",
// 			inputBody: `{"month":11,"year":1400}`,
// 			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
// 			},
// 			expectedStatusCode:  http.StatusBadRequest,
// 			expectedRequestBody: `{"message":"year: неверно указан год."}`,
// 		},
// 	}

// 	for _, testCase := range testTable {
// 		t.Run(testCase.name, func(t *testing.T) {
// 			c := gomock.NewController(t)
// 			defer c.Finish()

// 			control := mock_service.NewMockControl(c)
// 			testCase.mockBehavior(control, testCase.inputRequestReport)

// 			services := &service.Service{Control: control}
// 			h := NewHandler(services)

// 			r := mux.NewRouter()
// 			r.HandleFunc("/report", h.createReport).Methods("POST")

// 			w := httptest.NewRecorder()
// 			req := httptest.NewRequest("POST", "/report",
// 				bytes.NewBufferString(testCase.inputBody))

// 			r.ServeHTTP(w, req)

// 			assert.Equal(t, testCase.expectedStatusCode, w.Code)
// 			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
// 		})
// 	}
// }
