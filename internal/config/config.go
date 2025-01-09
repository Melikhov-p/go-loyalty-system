package config

import (
	"flag"
	"os"
	"time"
)

type WorkerConfig struct {
	PingInterval time.Duration
}

type configDB struct {
	DatabaseURI    string
	MigrationPath  string
	SecretKey      string
	ContextTimeout time.Duration
}

type Config struct {
	RunAddr       string
	AccrualAddr   string
	LogLevel      string
	DB            *configDB
	Worker        *WorkerConfig
	TokenLifeTime time.Duration
}

const (
	defaultRunAddr          = "localhost:8081"
	defaultDatabaseURI      = "_"
	defaultAccrualAddr      = "http://localhost:8080"
	defaultLogLevel         = "DEBUG"
	defaultMigrationPath    = "./internal/migrations/"
	defaultDBContextTimeout = 15 * time.Second
	defaultTokenLifeTime    = 24 * time.Hour
	defaultWorkerPingTasks  = 500 * time.Millisecond
)

func BuildConfig() *Config {
	cfg := Config{
		RunAddr:       defaultRunAddr,
		AccrualAddr:   defaultAccrualAddr,
		LogLevel:      defaultLogLevel,
		TokenLifeTime: defaultTokenLifeTime,
		DB: &configDB{
			DatabaseURI:    defaultDatabaseURI,
			MigrationPath:  defaultMigrationPath,
			SecretKey:      "3fac1504251a027465981346fb5b0d57d398e4df4a03253a4c7d1926e40e9907",
			ContextTimeout: defaultDBContextTimeout,
		},
		Worker: &WorkerConfig{
			PingInterval: defaultWorkerPingTasks,
		},
	}

	cfg.parseFlags()

	// Если флаг не установлен проверяем переменные окружения
	if cfg.RunAddr == defaultRunAddr {
		if osv, ok := os.LookupEnv("RUN_ADDRESS"); ok {
			cfg.RunAddr = osv
		}
	}
	if cfg.DB.DatabaseURI == defaultDatabaseURI {
		if osv, ok := os.LookupEnv("DATABASE_URI"); ok {
			cfg.DB.DatabaseURI = osv
		}
	}
	if cfg.AccrualAddr == defaultAccrualAddr {
		if osv, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
			cfg.AccrualAddr = osv
		}
	}

	return &cfg
}

func (c *Config) parseFlags() {
	flag.StringVar(&c.RunAddr, "a", defaultRunAddr, "Server host and port")
	flag.StringVar(&c.DB.DatabaseURI, "d", defaultDatabaseURI, "Database URI")
	flag.StringVar(&c.AccrualAddr, "r", defaultAccrualAddr, "Accrual system host and port")
	flag.StringVar(&c.LogLevel, "l", defaultLogLevel, "Logging level")
	flag.Parse()
}
