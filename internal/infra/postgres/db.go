package postgres

// Deprecated: replaced by GORM implementation in init.go and user_repository.go
// This file keeps a small compatibility wrapper to avoid compile errors in other packages.

import (
	"gorm.io/gorm"
)

// InitDB is a wrapper to the GORM initializer. Keep for compatibility with previous imports.
func InitDBWrapper(dsn string) (*gorm.DB, error) {
	return InitDB(dsn)
}
