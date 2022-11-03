package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"userbalance/internal/models"
)

type ControlPosgres struct {
	DB *sql.DB
}

func NewControlPostgres(db *sql.DB) *ControlPosgres {
	return &ControlPosgres{DB: db}
}

func (m *ControlPosgres) GetBalance(userId int) (models.User, error) {
	var balance int
	var id int
	rows, err := m.DB.Query("SELECT id, balance FROM users WHERE id = $1", userId)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &balance)
		if err != nil {
			return models.User{}, err
		}
	}

	return models.User{Id: id, Balance: balance}, err
}

func (m *ControlPosgres) ReplenishmentBalance(userId int, amount int, date string) error {
	check, err := m.checkUser(userId)
	if err != nil {
		return err
	}
	if check {
		tx, err := m.DB.Begin()
		if err != nil {
			return err
		}

		{
			stmt, err := tx.Prepare(`UPDATE users SET balance = balance + $1 WHERE id = $2;`)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer stmt.Close()

			if _, err := stmt.Exec(amount, userId); err != nil {
				tx.Rollback()
				return err
			}
		}

		{
			stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer stmt.Close()

			if _, err := stmt.Exec(userId, date, amount, "Пополнение баланса"); err != nil {
				tx.Rollback()
				return err
			}
		}
		return tx.Commit()

	} else {
		tx, err := m.DB.Begin()
		if err != nil {
			return err
		}

		{
			stmt, err := tx.Prepare(`INSERT INTO users (id, balance) VALUES ($1, $2);`)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer stmt.Close()

			if _, err := stmt.Exec(userId, amount); err != nil {
				tx.Rollback()
				return err
			}
		}

		{
			stmt, err := tx.Prepare(`INSERT INTO money_reserve_accounts (user_id) VALUES ($1);`)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer stmt.Close()

			if _, err := stmt.Exec(userId); err != nil {
				tx.Rollback()
				return err
			}
		}

		{
			stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
			if err != nil {
				tx.Rollback()
				return err
			}
			defer stmt.Close()

			if _, err := stmt.Exec(userId, date, amount, "Пополнение баланса"); err != nil {
				tx.Rollback()
				return err
			}
		}

		return tx.Commit()
	}
}

func (m *ControlPosgres) Transfer(fromUserId int, toUserId int, amount int, date string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	{
		stmt, err := tx.Prepare(`UPDATE users SET balance = balance - $1 WHERE id = $2;`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(amount, fromUserId); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(fromUserId, date, amount, fmt.Sprintf("Перевод средств пользователю %d", toUserId)); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`UPDATE users SET balance = balance + $1 WHERE id = $2;`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(amount, toUserId); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(toUserId, date, amount, fmt.Sprintf("Перевод средств от пользователя %d", fromUserId)); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (m *ControlPosgres) Reservation(userId int, serviceId int, orderId int, amount int, date string) error {
	var err error
	var service string

	if service, err = m.serviceTitle(serviceId); err != nil {
		return err
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	{
		stmt, err := tx.Prepare(`UPDATE users SET balance = balance - $1 WHERE id = $2`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(amount, userId); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`UPDATE money_reserve_accounts SET balance = balance + $1 WHERE user_id = $2`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(amount, userId); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO money_reserve_details (user_id, service_id, order_id, amount, date) VALUES ($1, $2, $3, $4, $5);`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(userId, serviceId, orderId, amount, date); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(userId, date, amount, fmt.Sprintf("Заказ №%d, услуга \"%s\"", orderId, service)); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (m *ControlPosgres) CancelReservation(userId int, serviceId int, orderId int, amount int, date string) error {
	var err error
	var service string

	if service, err = m.serviceTitle(serviceId); err != nil {
		return err
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	{
		stmt, err := tx.Prepare(`
			DELETE FROM money_reserve_details 
			WHERE user_id = $1 
			AND service_id = $2 
			AND order_id = $3
			AND amount = $4
			AND date = $5;`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		var result sql.Result
		if result, err = stmt.Exec(userId, serviceId, orderId, amount, date); err != nil {
			tx.Rollback()
			return err
		}
		if r, _ := result.RowsAffected(); r == 0 {
			tx.Rollback()
			return errors.New("по указанным критериям не было резерва")
		}
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO logs (user_id, date, amount, description) VALUES ($1, $2, $3, $4);`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(userId, date, amount, fmt.Sprintf("Отмена заказа №%d, услуга \"%s\"", orderId, service)); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`UPDATE users SET balance = balance + $1 WHERE id = $2`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err = stmt.Exec(amount, userId); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`UPDATE money_reserve_accounts SET balance = balance - $1 WHERE user_id = $2`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(amount, userId); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (m *ControlPosgres) Confirmation(userId int, serviceId int, orderId int, amount int, date string) error {
	var err error

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	{
		stmt, err := tx.Prepare(`
		UPDATE money_reserve_accounts 
		SET balance = balance - $1 
		WHERE user_id = (
			SELECT md.user_id 
			FROM money_reserve_details md 
			WHERE	md.user_id = $2 
			AND md.service_id = $3 
			AND	md.order_id = $4 
			AND	md.amount = $1
			AND md.date = $5
			)`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		var result sql.Result
		if result, err = stmt.Exec(amount, userId, serviceId, orderId, date); err != nil {
			tx.Rollback()
			return err
		}
		if r, _ := result.RowsAffected(); r == 0 {
			tx.Rollback()
			return errors.New("по указанным критериям не было резерва")
		}
	}

	{
		stmt, err := tx.Prepare(`
		DELETE FROM money_reserve_details 
		WHERE	user_id = $2 
		AND service_id = $3 
		AND order_id = $4 
		AND amount = $1
		AND date = $5
		`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(amount, userId, serviceId, orderId, date); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO report (user_id, service_id, amount, date) VALUES ($1, $2, $3, $4);`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(userId, serviceId, amount, date); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (m *ControlPosgres) CreateReport(fromDate string, toDate string) (map[string]int, error) {
	var report map[string]int = map[string]int{}

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
		var sum int
		err := rows.Scan(&title, &sum)
		if err != nil {
			return report, err
		}
		report[title] = sum
	}
	return report, err
}

func (m *ControlPosgres) GetHistory(userId int) ([]models.History, error) {
	var history []models.History = make([]models.History, 0)

	rows, err := m.DB.Query(`
		SELECT date, amount, description
		FROM logs
		WHERE user_id = $1
	`, userId)
	if err != nil {
		return history, err
	}

	defer rows.Close()

	for rows.Next() {
		var date time.Time
		var description string
		var amount int
		err := rows.Scan(&date, &amount, &description)
		if err != nil {
			return history, err
		}
		h := models.History{
			Date:        date.Format("02/01/2006"),
			Amount:      amount,
			Description: description,
		}
		history = append(history, h)
	}

	return history, err
}

func (m *ControlPosgres) checkUser(userId int) (bool, error) {
	var check bool
	var id int

	tx, err := m.DB.Begin()
	if err != nil {
		return check, err
	}

	{
		rows, err := tx.Query(`SELECT id FROM users WHERE id = $1`, userId)
		if err != nil {
			tx.Rollback()
			return check, err
		}

		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&id)
			if err != nil {
				tx.Rollback()
				return check, err
			}
		}

		if id != 0 {
			check = true
		}
		return check, tx.Commit()
	}
}

func (m *ControlPosgres) serviceTitle(serviceId int) (string, error) {
	var title string

	tx, err := m.DB.Begin()
	if err != nil {
		return title, err
	}

	rows, err := tx.Query("SELECT title FROM services WHERE id = $1", serviceId)
	if err != nil {
		tx.Rollback()
		return title, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&title)
		if err != nil {
			tx.Rollback()
			return title, err
		}
	}

	if title == "" {
		return title, errors.New("услуга не найдена")
	}

	return title, tx.Commit()
}
