package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config — конфигурация приложения из переменных окружения
type Config struct {
	HTTPAddr  string
	DBDSN     string
	AuthToken string // Bearer-токен (пусто = без аутентификации)
	LogLevel  string
}

// Load загружает конфигурацию из .env (если есть) и переменных окружения
func Load() (*Config, error) {
	// Игнорируем ошибку, если .env не найден
	_ = godotenv.Load()

	cfg := &Config{
		HTTPAddr:  getEnv("HTTP_ADDR", ":8080"),
		DBDSN:     getEnv("DB_DSN", ""),
		AuthToken: getEnv("AUTH_TOKEN", ""),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
	}

	if cfg.DBDSN == "" {
		// Собираем DSN из отдельных переменных
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "carservice")
		pass := getEnv("DB_PASS", "carservice")
		name := getEnv("DB_NAME", "carservice")
		sslmode := getEnv("DB_SSLMODE", "disable")
		cfg.DBDSN = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, pass, name, sslmode,
		)
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
