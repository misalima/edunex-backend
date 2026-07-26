package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Config struct {
	Port               string
	DBURL              string
	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseBucket     string
	SupabaseAnonKey    string
	SupabaseJWTKX      string
	SupabaseJWTKY      string
	GroqAPIKey         string
	GroqAPIURL         string
	GroqModel          string
	GroqTimeout        time.Duration
}

func Load() *Config {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "edunex")

	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode,
	)

	groqTimeoutStr := getEnv("GROQ_TIMEOUT", "60s")
	groqTimeout, err := time.ParseDuration(groqTimeoutStr)
	if err != nil {
		groqTimeout = 60 * time.Second
	}

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DBURL:              dbURL,
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseBucket:     getEnv("SUPABASE_BUCKET", ""),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseJWTKX:      getEnv("SUPABASE_JWT_K_X", ""),
		SupabaseJWTKY:      getEnv("SUPABASE_JWT_K_Y", ""),
		GroqAPIKey:         getEnv("GROQ_API_KEY", ""),
		GroqAPIURL:         getEnv("GROQ_API_URL", ""),
		GroqModel:          getEnv("GROQ_MODEL", ""),
		GroqTimeout:        groqTimeout,
	}

	validateRequired(cfg)

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue != "" {
		return defaultValue
	}
	return ""
}

func validateRequired(cfg *Config) {
	required := map[string]string{
		"SUPABASE_URL":      cfg.SupabaseURL,
		"SUPABASE_JWT_K_X":  cfg.SupabaseJWTKX,
		"SUPABASE_JWT_K_Y":  cfg.SupabaseJWTKY,
		"SUPABASE_ANON_KEY": cfg.SupabaseAnonKey,
		"DB_URL":            cfg.DBURL,
		"GROQ_API_KEY":      cfg.GroqAPIKey,
	}

	var missing []string
	for name, value := range required {
		if value == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("Missing environment variables: %v", missing)
	}
}
