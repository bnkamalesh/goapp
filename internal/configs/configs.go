package configs

import (
	"os"
	"strings"
	"time"

	"github.com/bnkamalesh/goapp/cmd/server/http"
	"github.com/bnkamalesh/goapp/internal/pkg/postgres"
)

type env string

func (e env) String() string {
	return string(e)
}

const (
	EnvLocal      env = "local"
	EnvTest       env = "test"
	EnvStaging    env = "staging"
	EnvProduction env = "production"
)

// Configs struct handles all dependencies required for handling configurations
type Configs struct {
	Environment env
	AppName     string
	AppVersion  string
}

// HTTP returns the configuration required for HTTP package
func (cfg *Configs) HTTP() (*http.Config, error) {
	return &http.Config{
		EnableAccessLog:   (cfg.Environment == EnvLocal) || (cfg.Environment == EnvTest),
		TemplatesBasePath: strings.TrimSpace(os.Getenv("TEMPLATES_BASEPATH")),
		Port:              8080,
		ReadTimeout:       time.Second * 5,
		WriteTimeout:      time.Second * 5,
		DialTimeout:       time.Second * 3,
	}, nil
}

func (cfg *Configs) Postgres() *postgres.Config {
	return &postgres.Config{
		Host:   os.Getenv("POSTGRES_HOST"),
		Port:   os.Getenv("POSTGRES_PORT"),
		Driver: "postgres",

		StoreName: os.Getenv("POSTGRES_STORENAME"),
		Username:  os.Getenv("POSTGRES_USERNAME"),
		Password:  os.Getenv("POSTGRES_PASSWORD"),

		ConnPoolSize: 24,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 6,
		IdleTimeout:  time.Minute,
		DialTimeout:  time.Second * 3,
	}
}

func (cfg *Configs) UserPostgresTable() string {
	return "users"
}

func (cfg *Configs) UserNotesPostgresTable() string {
	return "user_notes"
}

func loadEnv() env {
	switch env(os.Getenv("ENV")) {
	case EnvLocal:
		return EnvLocal
	case EnvTest:
		return EnvTest
	case EnvStaging:
		return EnvProduction
	case EnvProduction:
		return EnvProduction
	default:
		return EnvLocal
	}
}

// New returns an instance of Config with all the required dependencies initialized
func New() (*Configs, error) {
	return &Configs{
		Environment: loadEnv(),
		AppName:     os.Getenv("APP_NAME"),
		AppVersion:  os.Getenv("APP_VERSION"),
	}, nil
}
