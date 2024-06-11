package repo

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

func GetDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		url.QueryEscape(common.Config.DBUser),
		url.QueryEscape(common.Config.DBPass),
		common.Config.DBHost,
		common.Config.DBPort,
		url.QueryEscape(common.Config.DBName))
}

func SyncDBStructInsertionAndMigrations() error {
	slog.Info("running db migrations")
	m, err := migrate.New(
		"file://./db/migrations", GetDBURL()+"?sslmode=disable")
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("no changes in migrations")
			return nil
		}
		return fmt.Errorf("migrate.Up: %w", err)
	}

	return nil
}
