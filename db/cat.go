package db

import (
	"time"

	"gorm.io/gorm"
)

// CatID is the type for Cat record's primary key and unique identifier.
type CatID string

// SplotchID is the identifier of the OG, Splotch `Pat Junkie` the Cat.
const SplotchID CatID = "splotch"

// Cat represents a database record about a single cat.
type Cat struct {
	ID        CatID     // Unique identifier of the cat, must be URL-safe.
	Name      string    // Human-readable name of the cat.
	Pats      uint64    // Total number of pats received by the cat.
	LatestPat time.Time // Time when the latest pat was received.
}

// CatByID queries the cat with the given ID from the database.
func CatByID(tx *gorm.DB, id CatID) (Cat, error) {
	var c Cat
	result := tx.First(&c, "id = ?", id)
	if result.Error != nil {
		return Cat{}, result.Error
	}
	return c, nil
}
