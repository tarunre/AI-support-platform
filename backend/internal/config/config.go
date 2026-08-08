package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port             string
	DatabaseURL      string
	Environment      string
	JWTAccessSecret  string
	JWTRefreshSecret string

	// AI configuration
	GeminiAPIKey     string
	GeminiModel      string
	GroqAPIKey       string
	GroqModel        string
	AITimeout        int // seconds
	AIMaxRetries     int
	MaxReplyTokens   int
	ReplyTemperature float64

	// Queue / Worker configuration
	RedisURL        string
	WorkerCount     int
	QueueName       string
	MaxRetries      int
	RetryDelay      int // base seconds for exponential backoff
	WebSocketOrigin string

	// Email configuration
	EmailPollInterval int    // seconds between IMAP polls (default 60)
	MaxEmailRetries   int    // max SMTP retry attempts (default 3)
	AttachmentPath    string // local attachment storage directory

	// Analytics configuration
	ReportRetentionDays    int    // days to keep generated report files (default 30)
	MetricsRefreshInterval int    // seconds for scheduler interval (default 3600)
	AggregationInterval    int    // alias for MetricsRefreshInterval
	ReportStoragePath      string // directory for report files (default ./storage/reports)

	// Integration configuration
	IntegrationPollInterval int // seconds between integration event polls (default 30)
	WebhookSecret           string

	// Portal configuration
	AppURL string // base URL for the frontend app (used in magic-link portal emails)
}

// Load reads environment variables (from .env in development) and returns a Config.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	// Load .env file if present; silently ignored in production
	_ = godotenv.Load()

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		Environment:      getEnv("APP_ENV", "development"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		GeminiModel:      getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		GroqAPIKey:       getEnv("GROQ_API_KEY", ""),
		GroqModel:        getEnv("GROQ_MODEL", "llama-3.3-70b-versatile"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	if cfg.JWTAccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET environment variable is required")
	}
	if cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET environment variable is required")
	}

	if v := getEnv("AI_TIMEOUT", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.AITimeout = n
		}
	}
	if cfg.AITimeout == 0 {
		cfg.AITimeout = 30
	}

	if v := getEnv("AI_MAX_RETRIES", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.AIMaxRetries = n
		}
	}
	if cfg.AIMaxRetries == 0 {
		cfg.AIMaxRetries = 2
	}

	if v := getEnv("MAX_REPLY_TOKENS", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxReplyTokens = n
		}
	}
	if cfg.MaxReplyTokens == 0 {
		cfg.MaxReplyTokens = 1024
	}

	if v := getEnv("REPLY_TEMPERATURE", ""); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f <= 1 {
			cfg.ReplyTemperature = f
		}
	}
	if cfg.ReplyTemperature == 0 {
		cfg.ReplyTemperature = 0.3
	}

	cfg.RedisURL = getEnv("REDIS_URL", "")
	cfg.WebSocketOrigin = getEnv("WEBSOCKET_ORIGIN", "http://localhost:5173")
	cfg.AppURL = getEnv("APP_URL", "http://localhost:5173")
	cfg.QueueName = getEnv("QUEUE_NAME", "ai_jobs")

	if v := getEnv("WORKER_COUNT", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.WorkerCount = n
		}
	}
	if cfg.WorkerCount == 0 {
		cfg.WorkerCount = 3
	}

	if v := getEnv("MAX_RETRIES", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.MaxRetries = n
		}
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	if v := getEnv("RETRY_DELAY", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RetryDelay = n
		}
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 5
	}

	if v := getEnv("EMAIL_POLL_INTERVAL", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.EmailPollInterval = n
		}
	}
	if cfg.EmailPollInterval == 0 {
		cfg.EmailPollInterval = 60
	}

	if v := getEnv("MAX_EMAIL_RETRIES", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.MaxEmailRetries = n
		}
	}
	if cfg.MaxEmailRetries == 0 {
		cfg.MaxEmailRetries = 3
	}

	cfg.AttachmentPath = getEnv("ATTACHMENT_PATH", "./storage/attachments")

	if v := getEnv("REPORT_RETENTION_DAYS", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ReportRetentionDays = n
		}
	}
	if cfg.ReportRetentionDays == 0 {
		cfg.ReportRetentionDays = 30
	}

	if v := getEnv("METRICS_REFRESH_INTERVAL", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MetricsRefreshInterval = n
		}
	}
	if cfg.MetricsRefreshInterval == 0 {
		cfg.MetricsRefreshInterval = 3600
	}
	cfg.AggregationInterval = cfg.MetricsRefreshInterval
	cfg.ReportStoragePath = getEnv("REPORT_STORAGE_PATH", "./storage/reports")

	if v := getEnv("INTEGRATION_POLL_INTERVAL", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.IntegrationPollInterval = n
		}
	}
	if cfg.IntegrationPollInterval == 0 {
		cfg.IntegrationPollInterval = 30
	}
	cfg.WebhookSecret = getEnv("WEBHOOK_SECRET", "")

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
