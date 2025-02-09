package event_db

import (
	"backend/config"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateDB(cfg *config.Config) error {
	m, err := migrate.New(
		"file://migrations",
		fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable&search_path=public", cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName))

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No migrations to apply")
			return nil
		}
		return err
	}

	return nil
}
