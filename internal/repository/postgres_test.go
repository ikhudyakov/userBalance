package repository

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
	"userbalance"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCheckUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int
	}

	type mockBehavior func(args args, id int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int
		want         bool
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			id:   1,
			want: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)

				mock.ExpectCommit()
			},
		},

		{
			name: "user not found",
			args: args{
				userid: 2,
			},
			id:   0,
			want: false,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)

				mock.ExpectCommit()
			},
		},
		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).RowError(1, errors.New("some error"))
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)

				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id)

			got, err := r.checkUser(
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

func TestServiceTitle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		serviceid int
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
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.serviceid).WillReturnRows(rows)

				mock.ExpectCommit()
			},
		},
		{
			name: "service not found",
			args: args{
				serviceid: 10,
			},
			title:   "",
			want:    "услуга не найдена",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.serviceid).WillReturnRows(rows)

				mock.ExpectCommit()
			},
		},
		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"title"}).AddRow(title).RowError(1, errors.New("some error"))
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.serviceid).WillReturnRows(rows)

				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.title)

			got, err := r.serviceTitle(
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

func TestGetHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int
	}

	type mockBehavior func(args args, date time.Time, amount int, description string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		date         time.Time
		amount       int
		description  string
		want         []userbalance.History
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			date:        time.Date(2022, 10, 20, 0, 0, 0, 0, time.Local),
			amount:      100,
			description: "Пополнение баланса",
			want: []userbalance.History{
				{
					Date:        "20/10/2022",
					Amount:      100,
					Description: "Пополнение баланса",
				},
			},
			mockBehavior: func(args args, date time.Time, amount int, description string) {

				rows := sqlmock.NewRows([]string{"date", "amount", "description"}).AddRow(date, amount, description)
				mock.ExpectQuery("SELECT date, amount, description FROM logs").WithArgs(args.userid).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, date time.Time, amount int, description string) {

				mock.ExpectQuery("SELECT date, amount, description FROM logs").WithArgs(args.userid).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.date, testCase.amount, testCase.description)

			got, err := r.GetHistory(
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

func TestCreateReport(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		fromDate string
		toDate   string
	}

	type mockBehavior func(args args, title string, sum int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		title        string
		sum          int
		want         map[string]int
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				fromDate: "2022-10-01",
				toDate:   "2022-10-31",
			},
			title: "Услуга №1",
			sum:   100,
			want:  map[string]int{"Услуга №1": 100},
			mockBehavior: func(args args, title string, sum int) {

				rows := sqlmock.NewRows([]string{"title", "amount"}).AddRow(title, sum)
				mock.ExpectQuery("SELECT(.*)").WithArgs(args.fromDate, args.toDate).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, title string, sum int) {

				mock.ExpectQuery("SELECT(.*)").WithArgs(args.fromDate, args.toDate).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.title, testCase.sum)

			got, err := r.CreateReport(
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

func TestGetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int
	}

	type mockBehavior func(args args, id, balance int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int
		balance      int
		want         userbalance.User
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid: 1,
			},
			id:      1,
			balance: 100,
			want: userbalance.User{
				Id:      1,
				Balance: 100,
			},
			mockBehavior: func(args args, id, balance int) {

				rows := sqlmock.NewRows([]string{"id", "balance"}).AddRow(id, balance)
				mock.ExpectQuery("SELECT id, balance FROM users").WithArgs(args.userid).WillReturnRows(rows)
			},
		},

		{
			name:    "error",
			wantErr: true,
			mockBehavior: func(args args, id, balance int) {
				mock.ExpectQuery("SELECT id, balance FROM users").WithArgs(args.userid).WillReturnError(errors.New("some error"))
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id, testCase.balance)

			got, err := r.GetBalance(
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

func TestReplenishmentBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid int
		amount int
		date   string
		check  bool
	}

	type mockBehavior func(args args, id int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int
		wantErr      bool
	}{
		{
			name: "OK (user is exists)",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  true,
			},
			id: 1,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				if args.check {
					mock.ExpectBegin()
					mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WithArgs(args.userid, args.date, args.amount, "Пополнение баланса").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				} else {
					mock.ExpectBegin()
					mock.ExpectPrepare("INSERT INTO users").ExpectExec().WithArgs(args.userid, args.amount).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectPrepare("INSERT INTO money_reserve_accounts").ExpectExec().WithArgs(args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WithArgs(args.userid, args.date, args.amount, "Пополнение баланса").WillReturnResult(sqlmock.NewResult(1, 1))

					mock.ExpectCommit()
				}
			},
		},

		{
			name: "OK (new user)",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  false,
			},
			id: 0,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				if args.check {
					mock.ExpectBegin()
					mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WithArgs(args.userid, args.date, args.amount, "Пополнение баланса").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				} else {
					mock.ExpectBegin()
					mock.ExpectPrepare("INSERT INTO users").ExpectExec().WithArgs(args.userid, args.amount).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectPrepare("INSERT INTO money_reserve_accounts").ExpectExec().WithArgs(args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WithArgs(args.userid, args.date, args.amount, "Пополнение баланса").WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			},
		},

		{
			name: "error update",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  true,
			},
			id:      1,
			wantErr: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()

			},
		},

		{
			name: "error insert into logs if user exists",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  true,
			},
			id:      1,
			wantErr: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()

			},
		},

		{
			name: "error insert into users",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  false,
			},
			id:      0,
			wantErr: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()

			},
		},

		{
			name: "error insert into money_reserve_accounts",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  false,
			},
			id:      0,
			wantErr: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users").ExpectExec().WithArgs(args.userid, args.amount).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO money_reserve_accounts").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()

			},
		},

		{
			name: "error insert into logs in new user",
			args: args{
				userid: 1,
				amount: 100,
				date:   "2022-10-10",
				check:  false,
			},
			id:      0,
			wantErr: true,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("SELECT id FROM users").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users").ExpectExec().WithArgs(args.userid, args.amount).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO money_reserve_accounts").ExpectExec().WithArgs(args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()

			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id)

			err := r.ReplenishmentBalance(
				testCase.args.userid,
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

func TestReservation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid    int
		serviceId int
		orderId   int
		amount    int
		date      string
	}

	type mockBehavior func(args args, title string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		title        string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title: "Услуга №1",
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE money_reserve_accounts").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.userid, args.date, args.amount, fmt.Sprintf("Заказ №%d, услуга \"%s\"", args.orderId, title)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},

		{
			name: "service not found",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title).RowError(1, errors.New("услуга не найдена"))
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectRollback()
			},
		},

		{
			name: "error update users",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update money_reserve_accounts",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE money_reserve_accounts").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update money_reserve_details",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE money_reserve_accounts").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO money_reserve_details").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update logs",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE money_reserve_accounts").ExpectExec().WithArgs(args.amount, args.userid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.title)

			err := r.Reservation(
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

func TestConfirmation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid    int
		serviceId int
		orderId   int
		amount    int
		date      string
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
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE money_reserve_accounts").
					ExpectExec().WithArgs(args.amount, args.userid, args.serviceId, args.orderId, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.amount, args.userid, args.serviceId, args.orderId, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO report").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},

		{
			name: "error update money_reserve_accounts",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE money_reserve_accounts").
					ExpectExec().WithArgs(args.amount, args.userid, args.serviceId, args.orderId, args.date).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
		},

		{
			name: "error delete from money_reserve_details",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE money_reserve_accounts").
					ExpectExec().WithArgs(args.amount, args.userid, args.serviceId, args.orderId, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insert report",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE money_reserve_accounts").
					ExpectExec().WithArgs(args.amount, args.userid, args.serviceId, args.orderId, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.amount, args.userid, args.serviceId, args.orderId, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO report").WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			err := r.Confirmation(
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

func TestCancelReservation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		userid    int
		serviceId int
		orderId   int
		amount    int
		date      string
	}

	type mockBehavior func(args args, title string)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		title        string
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title: "Услуга №1",
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").
					ExpectExec().WithArgs(args.userid, args.date, args.amount, fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", args.orderId, title)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE users").
					ExpectExec().WithArgs(args.amount, args.userid).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE money_reserve_accounts").
					ExpectExec().WithArgs(args.amount, args.userid).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},

		{
			name: "error delete from money_reserve_details",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectRollback()
			},
		},

		{
			name: "error insert into from logs",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").
					ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update users",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").
					ExpectExec().WithArgs(args.userid, args.date, args.amount, fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", args.orderId, title)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE users").
					ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update money_reserve_accounts",
			args: args{
				userid:    1,
				serviceId: 1,
				orderId:   12,
				amount:    100,
				date:      "2022-10-10",
			},
			title:   "Услуга №1",
			wantErr: true,
			mockBehavior: func(args args, title string) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"title"}).AddRow(title)
				mock.ExpectQuery("SELECT title FROM services").WithArgs(args.userid).WillReturnRows(rows)
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectPrepare("DELETE FROM money_reserve_details").
					ExpectExec().WithArgs(args.userid, args.serviceId, args.orderId, args.amount, args.date).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").
					ExpectExec().WithArgs(args.userid, args.date, args.amount, fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", args.orderId, title)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE users").
					ExpectExec().WithArgs(args.amount, args.userid).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE money_reserve_accounts").
					ExpectExec().WillReturnError(errors.New("some error"))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.title)

			err := r.CancelReservation(
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

func TestTrannsfer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewControlPostgres(db)

	type args struct {
		fromuserid int
		touserid   int
		amount     int
		date       string
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
				fromuserid: 1,
				touserid:   1,
				amount:     100,
				date:       "2022-10-10",
			},
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.fromuserid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.fromuserid, args.date, args.amount, fmt.Sprintf("Перевод средств пользователю %d", args.touserid)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.touserid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.touserid, args.date, args.amount, fmt.Sprintf("Перевод средств от пользователя %d", args.touserid)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
		},

		{
			name: "error update fromuser balance",
			args: args{
				fromuserid: 1,
				touserid:   1,
				amount:     100,
				date:       "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.fromuserid).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
		},

		{
			name: "error into logs fromuser",
			args: args{
				fromuserid: 1,
				touserid:   1,
				amount:     100,
				date:       "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.fromuserid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.fromuserid, args.date, args.amount, fmt.Sprintf("Перевод средств пользователю %d", args.touserid)).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
		},

		{
			name: "error update touser balance",
			args: args{
				fromuserid: 1,
				touserid:   1,
				amount:     100,
				date:       "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.fromuserid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.fromuserid, args.date, args.amount, fmt.Sprintf("Перевод средств пользователю %d", args.touserid)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.touserid).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
		},

		{
			name: "error into logs touser",
			args: args{
				fromuserid: 1,
				touserid:   1,
				amount:     100,
				date:       "2022-10-10",
			},
			wantErr: true,
			mockBehavior: func(args args) {
				mock.ExpectBegin()
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.fromuserid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.fromuserid, args.date, args.amount, fmt.Sprintf("Перевод средств пользователю %d", args.touserid)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("UPDATE users").ExpectExec().WithArgs(args.amount, args.touserid).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO logs").ExpectExec().
					WithArgs(args.touserid, args.date, args.amount, fmt.Sprintf("Перевод средств от пользователя %d", args.touserid)).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args)

			err := r.Transfer(
				testCase.args.fromuserid,
				testCase.args.touserid,
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
