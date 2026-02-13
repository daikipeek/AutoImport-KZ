package config

import (
	"log"
	"os"
)

type Config struct {
	Addr         string
	DatabaseURL  string
	SessionKey   string
	SessionName  string
	StaticDir    string
	TemplateDir  string
	DefaultAdmin string
	DefaultPass  string
}

func Load() Config {
	cfg := Config{
		Addr:         getEnv("ADDR", ":8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/autoimport?sslmode=disable"),
		SessionKey:   getEnv("SESSION_KEY", "change-me-in-env"),
		SessionName:  getEnv("SESSION_NAME", "autoimport_session"),
		StaticDir:    getEnv("STATIC_DIR", "static"),
		TemplateDir:  getEnv("TEMPLATE_DIR", "internal/templates"),
		DefaultAdmin: getEnv("DEFAULT_ADMIN_EMAIL", "admin@demo.kz"),
		DefaultPass:  getEnv("DEFAULT_ADMIN_PASSWORD", "Admin123!"),
	}

	if cfg.SessionKey == "change-me-in-env" {
		log.Println("warning: SESSION_KEY is default; set a secure value in production")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
