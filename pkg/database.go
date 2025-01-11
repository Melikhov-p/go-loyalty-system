package pkg

import (
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DB.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("error open sql connection to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error ping database %w", err)
	}

	if err = makeMigrations(cfg, db); err != nil {
		return nil, fmt.Errorf("error making migrations %w", err)
	}

	return db, nil
}

func makeMigrations(cfg *config.Config, db *sql.DB) error {
	var err error

	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("error goose set dialect %w", err)
	}

	if err = goose.Up(db, cfg.DB.MigrationPath); err != nil {
		return fmt.Errorf("error up migrations %w", err)
	}

	return nil
}
