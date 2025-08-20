package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	AppPort         string
	PostgresDSN     string
	RedisAddr       string
	JWTSecret       string
	OTPTTL          time.Duration
	RateLimit       int
	RateLimitWindow time.Duration
}

func LoadConfig(logger *zap.Logger) *Config {
	if err := godotenv.Load("../../.env"); err != nil {
		logger.Info(".env not loaded; relying on environment variables", zap.Error(err))
	}
	cfg := &Config{
		AppPort:         mustEnv("APP_PORT", logger),
		PostgresDSN:     mustEnv("POSTGRES_DSN", logger),
		RedisAddr:       mustEnv("REDIS_ADDR", logger),
		JWTSecret:       mustEnv("JWT_SECRET", logger),
		OTPTTL:          mustDurationEnv("OTP_TTL", logger),
		RateLimit:       mustIntEnv("RATE_LIMIT", logger),
		RateLimitWindow: mustDurationEnv("RATE_LIMIT_WINDOW", logger),
	}

	return cfg
}

func mustEnv(key string, logger *zap.Logger) string {
	v := os.Getenv(key)
	if v == "" {
		logger.Fatal("missing required env var", zap.String("key", key))
	}
	return v
}

func mustDurationEnv(key string, logger *zap.Logger) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		logger.Fatal("missing required env var", zap.String("key", key))
	}

	duration, err := time.ParseDuration(v)
	if err != nil {
		logger.Fatal("invalid duration format",
			zap.String("key", key),
			zap.String("value", v),
			zap.String("expected_format", "use format like 2m, 10m, 1h"))
	}
	return duration
}

func mustIntEnv(key string, logger *zap.Logger) int {
	v := os.Getenv(key)
	if v == "" {
		logger.Fatal("missing required env var", zap.String("key", key))
	}

	value, err := strconv.Atoi(v)
	if err != nil {
		logger.Fatal("invalid integer format",
			zap.String("key", key),
			zap.String("value", v))
	}
	return value
}
