package middlewares

import (
	"database/sql"
	"net/http"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggerResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}

	Middleware struct {
		logger *zap.Logger
		cfg    *config.Config
		db     *sql.DB
	}
)

func NewMiddleware(logger *zap.Logger, cfg *config.Config, db *sql.DB) Middleware {
	return Middleware{
		logger: logger,
		cfg:    cfg,
		db:     db,
	}
}
