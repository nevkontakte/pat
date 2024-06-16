// Package dbtest contains database utilities for use in tests.
package dbtest

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InMemory create a disposable in-memory SQLite database.
//
// Each call returns a distinct, unique database, which is closed at the end of
// the test. See https://www.sqlite.org/inmemorydb.html.
func InMemory(t *testing.T) *gorm.DB {
	t.Helper()
	dbconn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create in-memory SQLite database: %s", err)
	}
	t.Cleanup(func() {
		sqlDB, err := dbconn.DB()
		if err != nil { // Should never happen.
			t.Fatalf("Failed to get underlying *sql.DB: %s", err)
		}
		sqlDB.Close()
	})
	return dbconn
}

// First calls tx.First() and fails the test if it fails.
func First(t *testing.T, tx *gorm.DB, dest any, conds ...any) {
	t.Helper()
	if result := tx.First(dest, conds...); result.Error != nil {
		t.Fatalf("Failed to fetch %T with %v: %s", dest, conds, result.Error)
	}
}

// Save calls tx.Save() and fails the test if it fails.
func Save(t *testing.T, tx *gorm.DB, values ...any) {
	t.Helper()
	for _, value := range values {
		if result := tx.Save(value); result.Error != nil {
			t.Fatalf("Failed to save %#v: %s", value, result.Error)
		}
	}
}
