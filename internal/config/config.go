package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// App holds server configuration from environment / .env
type App struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	WebDist      string
	LogLevel     string
}

// Notifier holds notifier microservice configuration
type Notifier struct {
	DatabaseURL  string
	CheckInterval time.Duration
	DedupeHours  int
}

// Load loads .env from current directory (if exists) and returns server config.
func Load() *App {
	_ = godotenv.Load()
	return &App{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/todo?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "todo-dev-secret-change-in-production"),
		WebDist:     getEnv("WEB_DIST", "./web/dist"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

// LoadNotifier loads .env and returns notifier config.
func LoadNotifier() *Notifier {
	_ = godotenv.Load()
	intervalMin := getEnvInt("NOTIFIER_INTERVAL_MINUTES", 10)
	dedupe := getEnvInt("NOTIFIER_DEDUPE_HOURS", 24)
	return &Notifier{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost/todo?sslmode=disable"),
		CheckInterval: time.Duration(intervalMin) * time.Minute,
		DedupeHours:  dedupe,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

// InitLog sets a consistent log format and optional prefix (e.g. "APP", "NOTIFIER").
func InitLog(prefix string) {
	if prefix != "" {
		log.SetPrefix("[" + prefix + "] ")
	}
	log.SetFlags(log.Ldate | log.Ltime)
}
