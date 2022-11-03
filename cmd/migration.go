package main

import (
	"fmt"
	c "userbalance/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func migration(conf *c.Config) error {
	m, err := migrate.New(
		"file://./migrations",
		fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable", conf.ConnectionType, conf.User, conf.Password, conf.Host, conf.DBPort, conf.DBname))
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}
	return err
}
