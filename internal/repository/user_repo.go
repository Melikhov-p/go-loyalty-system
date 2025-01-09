package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type UserRepo struct {
	logger *zap.Logger
	cfg    *config.Config
	db     *sql.DB
}

func (ur *UserRepo) NewEmptyUser() *models.User {
	return &models.User{
		ID:       -1,
		Login:    "",
		Password: "",
		AuthInfo: &models.AuthInfo{
			IsAuthenticated: false,
			Token:           "",
		},
		BalanceInfo: NewEmptyBalance(),
	}
}

func (ur *UserRepo) GetUserWithID(ctx context.Context, userID int) (*models.User, error) {
	query := `SELECT login FROM "user" WHERE id=$1`

	user := ur.NewEmptyUser()
	user.ID = userID

	ctxTimeout, cancel := context.WithTimeout(ctx, ur.cfg.DB.ContextTimeout)
	defer cancel()

	row := ur.db.QueryRowContext(ctxTimeout, query, userID)
	if err := row.Scan(&user.Login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserIDNotFound
		}
		return nil, fmt.Errorf("error scanning rows for user %w", err)
	}
	user.AuthInfo.IsAuthenticated = true

	return user, nil
}

func (ur *UserRepo) AddUser(ctx context.Context, login string, passHash string) (*models.User, error) {
	query := `INSERT INTO "user" (login, password) VALUES ($1, $2) RETURNING id`

	ctxTimeout, cancel := context.WithTimeout(ctx, ur.cfg.DB.ContextTimeout)
	defer cancel()

	tx, err := ur.db.Begin()
	defer func() {
		_ = tx.Rollback()
	}()

	row := tx.QueryRowContext(ctxTimeout, query, login, passHash)

	user := ur.NewEmptyUser()
	user.Login = login

	var id int64
	if err = row.Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserWithLoginExist
		}
		return nil, fmt.Errorf("error getting last insert id of user %w", err)
	}

	user.ID = int(id)

	_ = tx.Commit()
	return user, nil
}

func (ur *UserRepo) GetUserWithLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id, password FROM "user" WHERE login=$1`

	ctxWithTimeout, cancel := context.WithTimeout(ctx, ur.cfg.DB.ContextTimeout)
	defer cancel()

	user := ur.NewEmptyUser()
	row := ur.db.QueryRowContext(ctxWithTimeout, query, login)
	if err := row.Scan(&user.ID, &user.Password); err != nil {
		return nil, fmt.Errorf("error scanning row for login user %w", err)
	}

	return user, nil
}

func (ur *UserRepo) DeleteUserWithID(ctx context.Context, id int) error {
	query := `DELETE FROM "user" WHERE id=$1`

	ctxWithTimeout, cancel := context.WithTimeout(ctx, ur.cfg.DB.ContextTimeout)
	defer cancel()

	_, err := ur.db.ExecContext(ctxWithTimeout, query, id)
	if err != nil {
		return fmt.Errorf("error executing context for delete user %w", err)
	}

	return nil
}

func NewUserRepo(logger *zap.Logger, cfg *config.Config, db *sql.DB) *UserRepo {
	return &UserRepo{
		logger: logger,
		cfg:    cfg,
		db:     db,
	}
}
