package handlers

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/logger"
	"github.com/Melikhov-p/go-loyalty-system/internal/middlewares"
	"github.com/Melikhov-p/go-loyalty-system/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name         string
	body         string
	expectedCode int
	expectedBody string
}

var (
	server          *httptest.Server
	db              *sql.DB
	cfg             *config.Config
	testUserID      int
	testOrderNumber string
)

func TestHandlers(t *testing.T) {
	getServer(t)
	testUserID = 1
	testOrderNumber = "513"

	t.Run("USER", func(t *testing.T) {
		UserTestHandlers(t)
	})
	t.Run("ORDERS", func(t *testing.T) {
		OrderTestHandlers(t)
	})
	t.Run("BALANCE", func(t *testing.T) {
		BalanceTestHandlers(t)
	})
}

func getServer(t *testing.T) {
	cfg = config.BuildConfig()
	cfg.DB.MigrationPath = "../migrations"
	cfg.DB.DatabaseURI = "host=localhost user=postgres password=recnbr dbname=loyalty-system sslmode=disable"

	lgr, err := logger.BuildLogger(cfg.LogLevel)
	assert.NoError(t, err)

	db, err = pkg.ConnectDB(cfg)
	assert.NoError(t, err)

	r := chi.NewRouter()

	mdlwr := middlewares.NewMiddleware(lgr, cfg, db)
	r.Use(
		mdlwr.WithLogging,
		mdlwr.WithAuth,
		mdlwr.GzipMiddleware,
	)

	handlers := SetupHandlers(lgr, cfg, db)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", handlers.ForUser.UserRegister)
			r.Post("/login", handlers.ForUser.UserLogin)
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", handlers.ForOrder.CreateOrder)
				r.Get("/", handlers.ForOrder.GetOrders)
			})
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", handlers.ForBalance.GetBalance)
				r.Post("/withdraw", handlers.ForBalance.RequestWithdraw)
			})
			r.Get("/withdrawals", handlers.ForBalance.GetWithdrawals)
		})
	})

	server = httptest.NewServer(r)
}

func delTestUser(db *sql.DB, login string) error {
	_, err := db.Exec(`DELETE FROM "user" WHERE login=$1`, login)
	return err
}

func delTestOrder(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM "order" WHERE number=$1`, testOrderNumber)
	return err
}
