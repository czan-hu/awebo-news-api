package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServicePort string
	APIPrefix   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string
	JWTTTL    time.Duration

	CodeTTL       time.Duration
	CodeResendTTL time.Duration

	MailDriver   string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	UploadsDir      string
	MaxAvatarSizeMB int64
}

func Load() *Config {
	return &Config{
		ServicePort: getEnv("SERVICE_PORT", "8080"),
		APIPrefix:   getEnv("API_PREFIX", "/api/v1"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "app"),

		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:    time.Duration(getEnvInt("JWT_TTL_HOURS", 24*30)) * time.Hour,

		CodeTTL:       time.Duration(getEnvInt("CODE_TTL_SECONDS", 600)) * time.Second,
		CodeResendTTL: time.Duration(getEnvInt("CODE_RESEND_SECONDS", 30)) * time.Second,

		MailDriver:   getEnv("MAIL_DRIVER", "console"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "no-reply@awebo.example"),

		UploadsDir:      getEnv("UPLOADS_DIR", "uploads"),
		MaxAvatarSizeMB: int64(getEnvInt("MAX_AVATAR_SIZE_MB", 5)),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
