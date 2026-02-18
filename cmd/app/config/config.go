package config

import (
	"fmt"
	"os"

	"github.com/labstack/gommon/log"
)

type Config struct {
	Port               string
	DBURL              string
	JWTSecret          string
	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseBucket     string
}

func Load() *Config {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "edunex")

	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBURL:              dbURL,
		JWTSecret:          getEnv("JWT_SECRET", "secret-chave-muito-segura"),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseBucket:     getEnv("SUPABASE_BUCKET", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Info(fmt.Sprintf("Environment variable %s not set, using default value %s", key, defaultValue))
	return defaultValue
}
