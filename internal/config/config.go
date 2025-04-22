package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DB        DBConfig
	App       AppConfig
	JWT       JWTConfig
	Log       LogConfig
	API       APIConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AppConfig struct {
	Port string
	Env  string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type LogConfig struct {
	Level  string
	Format string
}

type APIConfig struct {
	Prefix             string
	CORSAllowedOrigins []string
}

type SecurityConfig struct {
	PasswordHashCost int
	SessionTimeout   int
}

type RateLimitConfig struct {
	Requests int
	Window   int
}

func Load() (*Config, error) {
	// Load JWT expiration duration
	jwtExpiration, err := time.ParseDuration(getEnvOrDefault("JWT_EXPIRATION", "24h"))
	if err != nil {
		return nil, err
	}

	// Load password hash cost
	passwordHashCost, err := strconv.Atoi(getEnvOrDefault("PASSWORD_HASH_COST", "10"))
	if err != nil {
		return nil, err
	}

	// Load session timeout
	sessionTimeout, err := strconv.Atoi(getEnvOrDefault("SESSION_TIMEOUT", "3600"))
	if err != nil {
		return nil, err
	}

	// Load rate limit config
	rateLimitRequests, err := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_REQUESTS", "100"))
	if err != nil {
		return nil, err
	}

	rateLimitWindow, err := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_WINDOW", "60"))
	if err != nil {
		return nil, err
	}

	return &Config{
		DB: DBConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5438"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			Name:     getEnvOrDefault("DB_NAME", "abi_banking"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		App: AppConfig{
			Port: getEnvOrDefault("APP_PORT", "8080"),
			Env:  getEnvOrDefault("APP_ENV", "development"),
		},
		JWT: JWTConfig{
			Secret:     getEnvOrDefault("JWT_SECRET", ""),
			Expiration: jwtExpiration,
		},
		Log: LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "debug"),
			Format: getEnvOrDefault("LOG_FORMAT", "text"),
		},
		API: APIConfig{
			Prefix:             getEnvOrDefault("API_PREFIX", "/api/v1"),
			CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
		},
		Security: SecurityConfig{
			PasswordHashCost: passwordHashCost,
			SessionTimeout:   sessionTimeout,
		},
		RateLimit: RateLimitConfig{
			Requests: rateLimitRequests,
			Window:   rateLimitWindow,
		},
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvList(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// Split by comma and trim spaces
	return strings.Split(value, ",")
}
