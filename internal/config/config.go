package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `json:"server"`
	Database   DatabaseConfig   `json:"database"`
	JWT        JWTConfig        `json:"jwt"`
	SMTP       SMTPConfig       `json:"smtp"`
	CBR        CBRConfig        `json:"cbr"`
	Encryption EncryptionConfig `json:"encryption"`
	RateLimit  RateLimitConfig  `json:"rate_limit"`
	API        APIConfig        `json:"api"`
	Log        LogConfig        `json:"log"`
	App        AppConfig        `json:"app"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret           string        `json:"secret"`
	ExpirationTime   time.Duration `json:"expiration_time"`
	RefreshDuration  time.Duration `json:"refresh_duration"`
	SigningAlgorithm string        `json:"signing_algorithm"`
}

// SMTPConfig represents SMTP configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	TLS      bool   `json:"tls"`
}

// CBRConfig represents Central Bank of Russia API configuration
type CBRConfig struct {
	BaseURL      string        `json:"base_url"`
	Timeout      time.Duration `json:"timeout"`
	RetryCount   int           `json:"retry_count"`
	RetryDelay   time.Duration `json:"retry_delay"`
	RateEndpoint string        `json:"rate_endpoint"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	CardDataKey     string `json:"card_data_key"`
	HMACSecret      string `json:"hmac_secret"`
	PGPPrivateKey   string `json:"pgp_private_key"`
	PGPPublicKey    string `json:"pgp_public_key"`
	KeyRotationDays int    `json:"key_rotation_days"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `json:"enabled"`
	RequestsPerHour int           `json:"requests_per_hour"`
	BurstSize       int           `json:"burst_size"`
	ExpiryTime      time.Duration `json:"expiry_time"`
}

// APIConfig represents API configuration
type APIConfig struct {
	Version            string   `json:"version"`
	Prefix             string   `json:"prefix"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level string `json:"level"`
}

// AppConfig represents application configuration
type AppConfig struct {
	Port string `json:"port"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		App: AppConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "bank",
			SSLMode:  "disable",
		},
		Log: LogConfig{
			Level: "info",
		},
		JWT: JWTConfig{
			ExpirationTime:   24 * time.Hour,
			RefreshDuration:  7 * 24 * time.Hour,
			SigningAlgorithm: "HS256",
		},
		RateLimit: RateLimitConfig{
			Enabled:         true,
			RequestsPerHour: 1000,
			BurstSize:       50,
			ExpiryTime:      1 * time.Hour,
		},
		API: APIConfig{
			Version:            "v1",
			Prefix:             "/api/v1",
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
	}
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

func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Override with environment variables if set
	cfg.Server.Host = getEnvOrDefault("SERVER_HOST", cfg.Server.Host)
	cfg.Server.Port = getEnvIntOrDefault("SERVER_PORT", cfg.Server.Port)
	cfg.Database.Host = getEnvOrDefault("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvIntOrDefault("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnvOrDefault("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnvOrDefault("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.DBName = getEnvOrDefault("DB_NAME", cfg.Database.DBName)
	cfg.Database.SSLMode = getEnvOrDefault("DB_SSL_MODE", cfg.Database.SSLMode)

	return cfg, nil
}
