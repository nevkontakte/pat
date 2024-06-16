package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Postgres connects to a PostgreSQL database.
func Postgres(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Bootstrap database state.
//
// Apply migrations and seed with initial data if missing. The operation is
// idempotent and should do nothing on an already set up database.
func Bootstrap(db *gorm.DB) error {
	if err := db.AutoMigrate(&Cat{}); err != nil {
		return fmt.Errorf("failed to auto-migrate data types: %w", err)
	}

	if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&Cat{
		ID:        SplotchID,
		Name:      "Splotch",
		Pats:      1,
		LatestPat: time.Now(),
	}); result.Error != nil {
		return fmt.Errorf("failed to create the initial record for Splotch: %s", result.Error)
	}
	return nil
}
