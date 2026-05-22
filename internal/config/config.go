package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	JWTSecret   string
	APIKey      string
	Environment string
	UploadDir   string
	BaseURL     string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SMTPFrom    string
	SMTPTo      string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "rmp_db"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		APIKey:      getEnv("API_KEY", ""),
		Environment: getEnv("ENV", "development"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		SMTPHost:    getEnv("SMTP_HOST", "smtpout.secureserver.net"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    getEnv("SMTP_USER", ""),
		SMTPPass:    getEnv("SMTP_PASS", ""),
		SMTPFrom:    getEnv("SMTP_FROM", ""),
		SMTPTo:      getEnv("SMTP_TO", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
