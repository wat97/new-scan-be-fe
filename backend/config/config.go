package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Server
	ServerPort string
	GinMode    string

	// CORS
	AllowedOrigins []string

	// Security
	RateLimitRequests int
	RateLimitDuration time.Duration
	MaxBodySize       int64

	// Logging
	LogLevel          string
	EnableSecurityLog bool
}

func LoadConfig() *Config {
	// Load .env file
	godotenv.Load()

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "root"),
		DBName:     getEnv("DB_NAME", "scandata"),

		// JWT
		JWTSecret:      getEnv("JWT_SECRET", "scandata-secret-key-2024"),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		// CORS
		AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", []string{"*"}),

		// Security
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration: getEnvDuration("RATE_LIMIT_DURATION", time.Minute),
		MaxBodySize:       getEnvInt64("MAX_BODY_SIZE", 10<<20), // 10MB

		// Logging
		LogLevel:          getEnv("LOG_LEVEL", "debug"),
		EnableSecurityLog: getEnvBool("ENABLE_SECURITY_LOG", true),
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) IsProduction() bool {
	return c.GinMode == "release"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
