package repository

import (
	"database/sql"
	"fmt"
	"time"
	"userbalance/pkg/api"

	sq "github.com/Masterminds/squirrel"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
)

type ControlPosgres struct {
	DB *sql.DB
}

func NewControlPostgres(db *sql.DB) *ControlPosgres {
	return &ControlPosgres{DB: db}
}

func (m *ControlPosgres) GetUser(userId int32) (*api.User, error) {
	var balance int32
	var id int32
	rows, err := m.DB.Query("SELECT id, balance FROM users WHERE id = $1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&id, &balance)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}

	return &api.User{Id: int32(id), Balance: int32(balance)}, err
}

func (m *ControlPosgres) GetUserForUpdate(tx *sql.Tx, userId int32) (*api.User, error) {
	var balance int32
	var id int32

	stmt, err := tx.Prepare(`SELECT id, balance FROM users WHERE id = $1 FOR UPDATE;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&id, &balance)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}

	return &api.User{Id: int32(id), Balance: int32(balance)}, err
}

func (m *ControlPosgres) GetReport(fromDate time.Time, toDate time.Time) (map[string]int32, error) {
	var report map[string]int32 = make(map[string]int32)

	rows, err := m.DB.Query(`
		SELECT s.title, SUM(r.amount) AS sumAmount
		FROM report r
		JOIN services s ON r.service_id = s.id
		WHERE r.date >= $1 AND r.date <= $2
		GROUP BY s.title
	`, fromDate, toDate)
	if err != nil {
		return report, err
	}

	defer rows.Close()

	for rows.Next() {
		var title string
		var sum int32
		err := rows.Scan(&title, &sum)
		if err != nil {
			return report, err
		}
		report[title] = sum
	}
	return report, err
}

func (m *ControlPosgres) GetHistory(requestHistory *api.RequestHistory) ([]*api.History, error) {
	var history []*api.History = make([]*api.History, 0)

	sql, args, err := sq.Select("date", "amount", "description").
		From("logs").
		Where(sq.Eq{"user_id": requestHistory.UserID}).
		OrderBy(fmt.Sprintf("%s %s", requestHistory.SortField, requestHistory.Direction)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := m.DB.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var date time.Time
		var description string
		var amount int32
		err := rows.Scan(&date, &amount, &description)
		if err != nil {
			return history, err
		}
		h := api.History{
			Date:        &timestamppb.Timestamp{Seconds: date.Unix()},
			Amount:      amount,
			Description: description,
		}
		history = append(history, &h)
	}

	return history, err
}

func (m *ControlPosgres) UpdateBalanceTx(tx *sql.Tx, userId int32, amount int32) error {

	stmt, err := tx.Prepare(`UPDATE users SET balance = $1 WHERE id = $2;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(amount, userId); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) InsertUserTx(tx *sql.Tx, userId int32, amount int32) error {

	stmt, err := tx.Prepare(`INSERT INTO users (id, balance) VALUES ($1, $2);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(userId, amount); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) InsertLogTx(tx *sql.Tx, userId int32, date time.Time, amount int32, description string) error {

	stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(userId, date, amount, description); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) InsertMoneyReserveAccountsTx(tx *sql.Tx, userId int32) error {

	stmt, err := tx.Prepare(`INSERT INTO money_reserve_accounts (user_id) VALUES ($1);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(userId); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) UpdateMoneyReserveAccountsTx(tx *sql.Tx, userId int32, amount int32) error {

	stmt, err := tx.Prepare(`UPDATE money_reserve_accounts SET balance = $1 WHERE user_id = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(amount, userId); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) GetBalanceReserveAccountsTx(tx *sql.Tx, userId int32) (int32, error) {
	var balance int32

	stmt, err := tx.Prepare(`SELECT balance FROM money_reserve_accounts WHERE user_id = $1 FOR UPDATE;`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&balance)
		if err != nil {
			return 0, err
		}
	}

	return balance, err
}

func (m *ControlPosgres) InsertMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int32, date time.Time) error {

	stmt, err := tx.Prepare(`INSERT INTO money_reserve_details (user_id, service_id, order_id, amount, date) VALUES ($1, $2, $3, $4, $5);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(userId, serviceId, orderId, amount, date); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) DeleteMoneyReserveDetailsTx(tx *sql.Tx, userId, serviceId, orderId, amount int32, date time.Time) (int64, error) {

	stmt, err := tx.Prepare(`
			DELETE FROM money_reserve_details 
			WHERE user_id = $1 
			AND service_id = $2 
			AND order_id = $3
			AND amount = $4
			AND date = $5;`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var result sql.Result
	if result, err = stmt.Exec(userId, serviceId, orderId, amount, date); err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (m *ControlPosgres) InsertReportTx(tx *sql.Tx, userId, serviceId, amount int32, date time.Time) error {

	stmt, err := tx.Prepare(`INSERT INTO report (user_id, service_id, amount, date) VALUES ($1, $2, $3, $4);`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(userId, serviceId, amount, date); err != nil {
		return err
	}
	return err
}

func (m *ControlPosgres) GetService(serviceId int32) (string, error) {
	var title string

	rows, err := m.DB.Query("SELECT title FROM services WHERE id = $1", serviceId)
	if err != nil {
		return title, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&title)
		if err != nil {
			return title, err
		}
	}

	return title, err
}
