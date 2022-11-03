package repository

import (
	"database/sql"
	"fmt"
	c "userbalance/internal/config"

	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func Connect(conf *c.Config) (*sql.DB, error) {
	var err error
	var conn string
	var db *sql.DB

	switch conf.ConnectionType {
	case "postgres":
		conn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", conf.Host, conf.DBPort, conf.User, conf.Password, conf.DBname)
		if db, err = sql.Open(conf.ConnectionType, conn); err != nil {
			return nil, err
		}
	case "mysql":
		cfg := mysql.Config{
			User:   conf.User,
			Passwd: conf.Password,
			Net:    "tcp",
			Addr:   fmt.Sprintf("%s:%d", conf.Host, conf.DBPort),
			DBName: conf.DBname,
		}
		if db, err = sql.Open(conf.ConnectionType, cfg.FormatDSN()); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid base type")
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
