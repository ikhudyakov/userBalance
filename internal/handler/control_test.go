package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
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
					models.User{
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
					models.User{}, errors.New("wrong userid"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"wrong userid"}`,
		},

		{
			name:      "error empty field",
			inputBody: `{}`,
			inputUser: models.User{},
			mockBehavior: func(s *mock_service.MockControl, user models.User) {
				s.EXPECT().GetBalance(user.Id).Return(
					models.User{}, errors.New("empty field"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"empty field"}`,
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
			inputBody: `{"userid":1,"amount":100,"serviceid":0,"orderid":0}`,
			inputTransaction: models.Transaction{
				UserID:    1,
				Amount:    100,
				ServiceID: 0,
				OrderID:   0,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().ReplenishmentBalance(&transaction).Return(nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"баланс пополнен"}`,
		},

		{
			name:      "error",
			inputBody: `{}`,
			inputTransaction: models.Transaction{
				UserID:    0,
				Amount:    0,
				ServiceID: 0,
				OrderID:   0,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().ReplenishmentBalance(&transaction).Return(errors.New("replenishment balance error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"replenishment balance error"}`,
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
			name:      "error",
			inputBody: `{}`,
			inputMoney: models.Money{
				FromUserID: 0,
				ToUserID:   0,
				Amount:     0,
				Date:       "",
			},
			mockBehavior: func(s *mock_service.MockControl, money models.Money) {
				s.EXPECT().Transfer(&money).Return(errors.New("transfer error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"transfer error"}`,
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
					Date:        "01/08/2022",
					Amount:      500,
					Description: "Пополнение баланса",
				}}, nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"entity":[{"date":"01/08/2022","amount":500,"description":"Пополнение баланса"}]}`,
		},

		{
			name:      "error",
			inputBody: `{"userid":0,"sortfield":"","direction":""}`,
			inputRequestHistory: models.RequestHistory{
				UserID:    0,
				SortField: "",
				Direction: "",
			},
			mockBehavior: func(s *mock_service.MockControl, requestHistory models.RequestHistory) {
				s.EXPECT().GetHistory(&requestHistory).Return(nil, errors.New("history error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"history error"}`,
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
			inputBody: `{"fromdate":"2022-08-01","todate":"2022-08-20"}`,
			inputRequestReport: models.RequestReport{
				FromDate: "2022-08-01",
				ToDate:   "2022-08-20",
			},
			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
				s.EXPECT().CreateReport(&requestReport).Return("localhost:8081/file/report.cvs", nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: `{"message":"localhost:8081/file/report.cvs"}`,
		},

		{
			name:      "error",
			inputBody: `{"fromdate":"","todate":""}`,
			inputRequestReport: models.RequestReport{
				FromDate: "",
				ToDate:   "",
			},
			mockBehavior: func(s *mock_service.MockControl, requestReport models.RequestReport) {
				s.EXPECT().CreateReport(&requestReport).Return("", errors.New("report error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"report error"}`,
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
			name:      "error",
			inputBody: `{"userid":0,"amount":0,"serviceid":0,"orderid":0}`,
			inputTransaction: models.Transaction{
				UserID:    0,
				Amount:    0,
				ServiceID: 0,
				OrderID:   0,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().Reservation(&transaction).Return(errors.New("reservstion error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"reservstion error"}`,
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
			name:      "error",
			inputBody: `{"userid":0,"amount":0,"serviceid":0,"orderid":0}`,
			inputTransaction: models.Transaction{
				UserID:    0,
				Amount:    0,
				ServiceID: 0,
				OrderID:   0,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().Confirmation(&transaction).Return(errors.New("confirmation error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"confirmation error"}`,
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
			name:      "error",
			inputBody: `{"userid":0,"amount":0,"serviceid":0,"orderid":0}`,
			inputTransaction: models.Transaction{
				UserID:    0,
				Amount:    0,
				ServiceID: 0,
				OrderID:   0,
			},
			mockBehavior: func(s *mock_service.MockControl, transaction models.Transaction) {
				s.EXPECT().CancelReservation(&transaction).Return(errors.New("cancel error"))
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedRequestBody: `{"message":"cancel error"}`,
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
