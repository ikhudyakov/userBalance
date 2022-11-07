package repository

import (
	"fmt"
	c "userbalance/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func Migration(conf *c.Config, up, down bool) error {
	m, err := migrate.New(
		"file://"+conf.MigrationPath,
		fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable",
			conf.ConnectionType,
			conf.User,
			conf.Password,
			conf.DBHost,
			conf.DBPort,
			conf.DBname))
	if err != nil {
		return err
	}

	if up {
		if err := m.Up(); err != nil {
			return err
		}
	}

	if down {
		if err := m.Down(); err != nil {
			return err
		}
	}
	return err
}
