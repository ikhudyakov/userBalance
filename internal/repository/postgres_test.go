package repository

import (
	"errors"
	"log"
	"testing"
	"time"
	"userbalance/pkg/api"

	"github.com/DATA-DOG/go-sqlmock"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
	}

	type mockBehavior func(args args, id, balance int32)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int32
		balance      int32
		want         *api.User
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			id:      1,
			balance: 100,
			want: &api.User{
				Id:      1,
				Balance: 100,
			},
			mockBehavior: func(args args, id, balance int32) {
				rows := sqlmock.NewRows([]string{"id", "balance"}).AddRow(id, balance)
				mock.ExpectQuery("SELECT id, balance FROM users").WithArgs(args.userid).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, id, balance int32) {
				mock.ExpectQuery("SELECT id, balance FROM users").WithArgs(args.userid).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id, testCase.balance)

			got, err := r.GetUser(
				testCase.args.userid)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestGetUserForUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
	}

	type mockBehavior func(args args, id, balance int32)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int32
		balance      int32
		want         *api.User
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			id:      1,
			balance: 100,
			want: &api.User{
				Id:      1,
				Balance: 100,
			},
			mockBehavior: func(args args, id, balance int32) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "balance"}).AddRow(id, balance)
				mock.ExpectPrepare("SELECT id, balance FROM users").ExpectQuery().WithArgs(args.userid).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, id, balance int32) {
				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT id, balance FROM users").ExpectQuery().WithArgs(args.userid).WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id, testCase.balance)

			tx, _ := db.Begin()
			got, err := r.GetUserForUpdate(
				tx,
				testCase.args.userid)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestGetReport(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		fromDate time.Time
		toDate   time.Time
	}

	type mockBehavior func(args args, title string, sum int32)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		title        string
		sum          int32
		want         map[string]int32
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				fromDate: time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
				toDate:   time.Date(2022, 11, 30, 0, 0, 0, 0, time.Local),
			},
			title: "Услуга №1",
			sum:   100,
			want:  map[string]int32{"Услуга №1": 100},
			mockBehavior: func(args args, title string, sum int32) {
				rows := sqlmock.NewRows([]string{"title", "amount"}).AddRow(title, sum)
				mock.ExpectQuery("SELECT(.*)").WithArgs(args.fromDate, args.toDate).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, title string, sum int32) {
				mock.ExpectQuery("SELECT(.*)").WithArgs(args.fromDate, args.toDate).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.title, testCase.sum)

			got, err := r.GetReport(
				testCase.args.fromDate, testCase.args.toDate)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestGetHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		requestHistory *api.RequestHistory
	}

	type mockBehavior func(args args, date time.Time, amount int32, description string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		date         timestamppb.Timestamp
		amount       int32
		description  string
		want         []*api.History
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				requestHistory: &api.RequestHistory{
					UserID:    1,
					SortField: "asc",
					Direction: "amount",
				},
			},
			date:        timestamppb.Timestamp{Seconds: time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local).Unix()},
			amount:      100,
			description: "Пополнение баланса",
			want: []*api.History{
				{
					Date:        &timestamppb.Timestamp{Seconds: time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local).Unix()},
					Amount:      100,
					Description: "Пополнение баланса",
				},
			},
			mockBehavior: func(args args, date time.Time, amount int32, description string) {
				rows := sqlmock.NewRows([]string{"date", "amount", "description"}).AddRow(date, amount, description)
				mock.ExpectQuery("SELECT date, amount, description FROM logs").WithArgs(args.requestHistory.UserID).WillReturnRows(rows)
			},
		},

		{
			name: "error",
			args: args{
				requestHistory: &api.RequestHistory{
					UserID:    1,
					SortField: "asc",
					Direction: "amount",
				},
			},
			wantErr: true,
			mockBehavior: func(args args, date time.Time, amount int32, description string) {
				mock.ExpectQuery("SELECT date, amount, description FROM logs").WithArgs(args.requestHistory.UserID).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.date.AsTime(), testCase.amount, testCase.description)

			got, err := r.GetHistory(
				testCase.args.requestHistory)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestUpdateBalanceTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
		amount int32
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
				amount: 100,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid: 1,
				amount: 100,
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnError(errors.New("error update"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.UpdateBalanceTx(
				tx,
				testCase.args.userid,
				testCase.args.amount)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsertUserTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
		amount int32
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
				amount: 100,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users").ExpectExec().WithArgs(args.userid, args.amount).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid: 1,
				amount: 100,
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users").ExpectExec().WithArgs(args.userid, args.amount).WillReturnError(errors.New("error insert"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.InsertUserTx(
				tx,
				testCase.args.userid,
				testCase.args.amount)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsertLogTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid      int32
		date        time.Time
		amount      int32
		description string
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid:      1,
				date:        time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
				amount:      100,
				description: "Пополнение баланса",
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WithArgs(args.userid, args.date, args.amount, args.description).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid:      1,
				date:        time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
				amount:      100,
				description: "Пополнение баланса",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WithArgs(args.userid, args.date, args.amount, args.description).WillReturnError(errors.New("error insert"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.InsertLogTx(
				tx,
				testCase.args.userid,
				testCase.args.date,
				testCase.args.amount,
				testCase.args.description)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsertMoneyReserveAccountsTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO money_reserve_accounts").ExpectExec().WithArgs(args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid: 1,
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO money_reserve_accounts").ExpectExec().WithArgs(args.userid).WillReturnError(errors.New("error insert"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.InsertMoneyReserveAccountsTx(
				tx,
				testCase.args.userid)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateMoneyReserveAccountsTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
		amount int32
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
				amount: 100,
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE money_reserve_accounts").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid: 1,
				amount: 100,
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE money_reserve_accounts").ExpectExec().WithArgs(args.amount, args.userid).WillReturnError(errors.New("error update"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.UpdateMoneyReserveAccountsTx(
				tx,
				testCase.args.userid,
				testCase.args.amount)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBalanceReserveAccountsTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int32
	}

	type mockBehavior func(args args, balance int32)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		balance      int32
		want         int32
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			balance: 100,
			want:    100,
			mockBehavior: func(args args, balance int32) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"balance"}).AddRow(balance)
				mock.ExpectPrepare("SELECT balance FROM money_reserve_accounts").ExpectQuery().WithArgs(args.userid).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, balance int32) {
				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT id, balance FROM users").ExpectQuery().WithArgs(args.userid).WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.balance)

			tx, _ := db.Begin()
			got, err := r.GetBalanceReserveAccountsTx(
				tx,
				testCase.args.userid)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestInsertMoneyReserveDetailsTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid    int32
		serviceId int32
		orderId   int32
		amount    int32
		date      time.Time
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   1,
				amount:    100,
				date:      time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO money_reserve_details").ExpectExec().WithArgs(
					args.userid,
					args.serviceId,
					args.orderId,
					args.amount,
					args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   1,
				amount:    100,
				date:      time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO money_reserve_details").ExpectExec().WithArgs(
					args.userid,
					args.serviceId,
					args.orderId,
					args.amount,
					args.date).
					WillReturnError(errors.New("error insert"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.InsertMoneyReserveDetailsTx(
				tx,
				testCase.args.userid,
				testCase.args.serviceId,
				testCase.args.orderId,
				testCase.args.amount,
				testCase.args.date)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteMoneyReserveDetailsTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid    int32
		serviceId int32
		orderId   int32
		amount    int32
		date      time.Time
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		want         int64
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   1,
				amount:    100,
				date:      time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			},
			want: 1,

			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").ExpectExec().WithArgs(
					args.userid,
					args.serviceId,
					args.orderId,
					args.amount,
					args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   1,
				amount:    100,
				date:      time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").ExpectExec().WithArgs(
					args.userid,
					args.serviceId,
					args.orderId,
					args.amount,
					args.date).
					WillReturnError(errors.New("error delete"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			got, err := r.DeleteMoneyReserveDetailsTx(
				tx,
				testCase.args.userid,
				testCase.args.serviceId,
				testCase.args.orderId,
				testCase.args.amount,
				testCase.args.date)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}

func TestInsertReportTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid    int32
		serviceId int32
		amount    int32
		date      time.Time
	}

	type mockBehavior func(args args)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid:    1,
				serviceId: 1,
				amount:    100,
				date:      time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO report").ExpectExec().WithArgs(
					args.userid,
					args.serviceId,
					args.amount,
					args.date).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name: "error",
			args: args{
				userid:    1,
				serviceId: 1,
				amount:    100,
				date:      time.Date(2022, 11, 01, 0, 0, 0, 0, time.Local),
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO report").ExpectExec().WithArgs(
					args.userid,
					args.serviceId,
					args.amount,
					args.date).WillReturnError(errors.New("error insert"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			tx, _ := db.Begin()
			err := r.InsertReportTx(
				tx,
				testCase.args.userid,
				testCase.args.serviceId,
				testCase.args.amount,
				testCase.args.date)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		serviceid int32
	}

	type mockBehavior func(args args, title string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		title        string
		want         string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				serviceid: 1,
			},
			title: "Услуга №1",
			want:  "Услуга №1",
			mockBehavior: func(args args, title string) {
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.serviceid).WillReturnRows(rows)
			},
		},

		{
			name: "error",
			args: args{
				serviceid: 1,
			},
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.serviceid).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.title)

			got, err := r.GetService(
				testCase.args.serviceid)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want, got)
			}
		})
	}
}
