package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"userbalance/internal/models"
	"userbalance/internal/service"
	mock_service "userbalance/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandler_getBalance(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, user models.User)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           models.User
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"userid":1}`,
			inputUser: models.User{
				Id: 1,
			},
			mockBehavior: func(s *mock_service.MockControl, user models.User) {
				s.EXPECT().GetBalance(user.Id).Return(
					&models.User{
						Id:      1,
						Balance: 100}, nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"userid":1,"balance":100}`,
		},

		{
			name:      "error wrong userid",
			inputBody: `{"userid":10}`,
			inputUser: models.User{
				Id: 10,
			},
			mockBehavior: func(s *mock_service.MockControl, user models.User) {
				s.EXPECT().GetBalance(user.Id).Return(
					nil, errors.New("wrong userid"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"wrong userid"}`,
		},

		{
			name:                "error empty field",
			inputBody:           `{}`,
			inputUser:           models.User{},
			mockBehavior:        func(s *mock_service.MockControl, user models.User) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"userid: id пользователя не может быть не указан либо \u003c= 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputUser)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/", h.getBalance).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_replenishmentBalance(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, replenishment models.Replenishment)

	testTable := []struct {
		name                string
		inputBody           string
		inputReplenishment  models.Replenishment
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"userid":1,"amount":100,"date":"2022-11-01"}`,
			inputReplenishment: models.Replenishment{
				UserID: 1,
				Amount: 100,
				Date:   "2022-11-01",
			},
			mockBehavior: func(s *mock_service.MockControl, replenishment models.Replenishment) {
				s.EXPECT().ReplenishmentBalance(&replenishment).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"баланс пополнен"}`,
		},

		{
			name:      "error userId <= 0",
			inputBody: `{"userid":-1,"amount":100,"date":"2022-11-01"}`,
			mockBehavior: func(s *mock_service.MockControl, replenishment models.Replenishment) {
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"userid: id пользователя не может быть \u003c= 0."}`,
		},

		{
			name:      "error amount <= 0",
			inputBody: `{"userid":1,"amount":-100,"date":"2022-11-01"}`,
			mockBehavior: func(s *mock_service.MockControl, replenishment models.Replenishment) {
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"amount: сумма пополнения должна быть больше 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputReplenishment)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/topup", h.replenishmentBalance).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/topup",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_transfer(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, money models.Money)

	testTable := []struct {
		name                string
		inputBody           string
		inputMoney          models.Money
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"fromuserid":1,"touserid":2,"amount":100,"date":"2022-08-01"}`,
			inputMoney: models.Money{
				FromUserID: 1,
				ToUserID:   2,
				Amount:     100,
				Date:       "2022-08-01",
			},
			mockBehavior: func(s *mock_service.MockControl, money models.Money) {
				s.EXPECT().Transfer(&money).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"перевод стредств выполнен"}`,
		},

		{
			name:                "error fromUserId <=0",
			inputBody:           `{"fromuserid":-1,"touserid":2,"amount":100,"date":"2022-08-01"}`,
			mockBehavior:        func(s *mock_service.MockControl, money models.Money) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"fromuserid: id пользователя не может быть \u003c= 0."}`,
		},

		{
			name:                "error toUserId <=0",
			inputBody:           `{"fromuserid":1,"touserid":-2,"amount":100,"date":"2022-08-01"}`,
			mockBehavior:        func(s *mock_service.MockControl, money models.Money) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"touserid: id пользователя не может быть \u003c= 0."}`,
		},

		{
			name:                "error toUserId == fromUserId",
			inputBody:           `{"fromuserid":1,"touserid":1,"amount":100,"date":"2022-08-01"}`,
			mockBehavior:        func(s *mock_service.MockControl, money models.Money) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"touserid: невозможно перевести самому себе."}`,
		},

		{
			name:                "error amount <=0",
			inputBody:           `{"fromuserid":1,"touserid":2,"amount":-100,"date":"2022-08-01"}`,
			mockBehavior:        func(s *mock_service.MockControl, money models.Money) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"amount: сумма перевода должна быть больше 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputMoney)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/transfer", h.transfer).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/transfer",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_getHistory(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, requestHistory models.RequestHistory)

	testTable := []struct {
		name                string
		inputBody           string
		inputRequestHistory models.RequestHistory
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"userid":1,"sortfield":"amount","direction":"desc"}`,
			inputRequestHistory: models.RequestHistory{
				UserID:    1,
				SortField: "amount",
				Direction: "desc",
			},
			mockBehavior: func(s *mock_service.MockControl, requestHistory models.RequestHistory) {
				s.EXPECT().GetHistory(&requestHistory).Return([]models.History{{
					Date:        time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
					Amount:      500,
					Description: "Пополнение баланса",
				}}, nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"entity":[{"date":"2022-11-01T00:00:00+03:00","amount":500,"description":"Пополнение баланса"}]}`,
		},

		{
			name:      "error userId <= 0",
			inputBody: `{"userid":0,"sortfield":"","direction":""}`,
			mockBehavior: func(s *mock_service.MockControl, requestHistory models.RequestHistory) {
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"userid: id пользователя не может быть \u003c= 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputRequestHistory)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/history", h.getHistory).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/history",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_createReport(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, requestReport models.RequestReport)

	testTable := []struct {
		name                string
		inputBody           string
		inputRequestReport  models.RequestReport
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"month":11,"year":2022}`,
			inputRequestReport: models.RequestReport{
				Month: 11,
				Year:  2022,
			},
			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
				s.EXPECT().CreateReport(&requestReport).Return("localhost:8081/file/report.cvs", nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"localhost:8081/file/report.cvs"}`,
		},

		{
			name:      "error wrong month",
			inputBody: `{"month":22,"year":2022}`,
			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"month: месяца не может быть \u003e 12."}`,
		},

		{
			name:      "error wrong year",
			inputBody: `{"month":11,"year":1400}`,
			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
			},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"year: неверно указан год."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputRequestReport)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/report", h.createReport).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/report",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_reservation(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, transaction models.Transaction)

	testTable := []struct {
		name                string
		inputBody           string
		inputTransaction    models.Transaction
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"userid":1,"amount":100,"serviceid":1,"orderid":12}`,
			inputTransaction: models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().Reservation(&transaction).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"резервирование средств прошло успешно"}`,
		},

		{
			name:                "error user <= 0",
			inputBody:           `{"userid":0,"amount":100,"serviceid":1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"userid: id пользователя не может быть не указан либо \u003c= 0."}`,
		},

		{
			name:                "error amount <= 0",
			inputBody:           `{"userid":1,"amount":0,"serviceid":1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"amount: стоимость услуги должна быть больше 0."}`,
		},

		{
			name:                "error serviceid <= 0",
			inputBody:           `{"userid":1,"amount":100,"serviceid":-1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"serviceid: id услуги не может быть \u003c= 0."}`,
		},

		{
			name:                "error orderid <= 0",
			inputBody:           `{"userid":1,"amount":100,"serviceid":1,"orderid":-1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"orderid: номер заказа не может быть \u003c= 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputTransaction)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/reserv", h.reservation).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/reserv",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_confirmation(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, transaction models.Transaction)

	testTable := []struct {
		name                string
		inputBody           string
		inputTransaction    models.Transaction
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"userid":1,"amount":100,"serviceid":1,"orderid":12}`,
			inputTransaction: models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().Confirmation(&transaction).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"средства из резерва были списаны успешно"}`,
		},

		{
			name:                "error user <= 0",
			inputBody:           `{"userid":0,"amount":100,"serviceid":1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"userid: id пользователя не может быть не указан либо \u003c= 0."}`,
		},

		{
			name:                "error amount <= 0",
			inputBody:           `{"userid":1,"amount":0,"serviceid":1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"amount: стоимость услуги должна быть больше 0."}`,
		},

		{
			name:                "error serviceid <= 0",
			inputBody:           `{"userid":1,"amount":100,"serviceid":-1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"serviceid: id услуги не может быть \u003c= 0."}`,
		},

		{
			name:                "error orderid <= 0",
			inputBody:           `{"userid":1,"amount":100,"serviceid":1,"orderid":-1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"orderid: номер заказа не может быть \u003c= 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputTransaction)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/confirm", h.confirmation).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/confirm",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_cancelReservation(t *testing.T) {

	type mockBehavior func(s *mock_service.MockControl, transaction models.Transaction)

	testTable := []struct {
		name                string
		inputBody           string
		inputTransaction    models.Transaction
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"userid":1,"amount":100,"serviceid":1,"orderid":12}`,
			inputTransaction: models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 1,
				OrderID:   12,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().CancelReservation(&transaction).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"разрезервирование средств прошло успешно"}`,
		},

		{
			name:                "error user <= 0",
			inputBody:           `{"userid":0,"amount":100,"serviceid":1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"userid: id пользователя не может быть не указан либо \u003c= 0."}`,
		},

		{
			name:                "error amount <= 0",
			inputBody:           `{"userid":1,"amount":0,"serviceid":1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"amount: стоимость услуги должна быть больше 0."}`,
		},

		{
			name:                "error serviceid <= 0",
			inputBody:           `{"userid":1,"amount":100,"serviceid":-1,"orderid":1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"serviceid: id услуги не может быть \u003c= 0."}`,
		},

		{
			name:                "error orderid <= 0",
			inputBody:           `{"userid":1,"amount":100,"serviceid":1,"orderid":-1}`,
			mockBehavior:        func(s *mock_service.MockControl, transaction models.Transaction) {},
			expectedStatusCode:  http.StatusBadRequest,
			expectedRequestBody: `{"message":"orderid: номер заказа не может быть \u003c= 0."}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			control := mock_service.NewMockControl(c)
			testCase.mockBehavior(control, testCase.inputTransaction)

			services := &service.Service{Control: control}
			h := NewHandler(services)

			r := mux.NewRouter()
			r.HandleFunc("/cancel", h.cancelReservation).Methods("POST")

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/cancel",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}
