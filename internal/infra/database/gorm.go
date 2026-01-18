package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes and returns a GORM DB using the provided DSN.
// Example DSN formats accepted by the postgres driver:
// - postgres://user:password@host:port/dbname
// - host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable
func InitDB(dsn string) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "[gorm] ", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm db: %w", err)
	}

	// configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql DB from gorm: %w", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}
