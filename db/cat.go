package db

import (
	"fmt"
	"time"

	"github.com/nevkontakte/pat/behavior"
	"github.com/nevkontakte/pat/chrono"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// CatID is the type for Cat record's primary key and unique identifier.
type CatID string

// Seed returns the noise seed for the cat, given noise type.
func (id CatID) Seed(cue string) []byte {
	return []byte(cue + string(id))
}

func (id CatID) Name() string {
	return cases.Title(language.AmericanEnglish).String(string(id))
}

// SplotchID is the identifier of the OG, Splotch `Pat Junkie` the Cat.
const SplotchID CatID = "splotch"

// Mood represents Cat's current state of mind.
type Mood string

const (
	MoodIdle      Mood = "idle"
	MoodIdleBlink Mood = "idle_blink"
	MoodIdleHappy Mood = "idle_happy"
	MoodImpatient Mood = "impatient"
	MoodPat       Mood = "pat"
	MoodSecret    Mood = "secret"
)

// Cat represents a database record about a single cat.
type Cat struct {
	ID        CatID     // Unique identifier of the cat, must be URL-safe.
	Name      string    // Human-readable name of the cat.
	Pats      uint64    // Total number of pats received by the cat.
	LatestPat time.Time // Time when the latest pat was received.
}

// Mood corresponding to Cat's current state.
func (c Cat) Mood() Mood {
	now := chrono.Now()
	sincePat := now.Sub(c.LatestPat)

	// Someone is petting the cat right now!
	if sincePat < 5*time.Second {
		return MoodPat
	}

	// Someone petted the cat recently, he's happy.
	if sincePat < 30*time.Minute {
		return MoodIdleHappy
	}

	// In the next three hours, the cat may remember getting petted and get happy again.
	moodSwing := behavior.Spread(-3*time.Hour, 3*time.Hour, c.noise(c.ID.Seed("happy"), 5*time.Minute).At(now))
	if sincePat+moodSwing < 30*time.Minute {
		return MoodIdleHappy
	}

	// Cat's just chillin'.
	if sincePat < 7*24*time.Hour {
		return MoodIdle
	}

	// It's been far too long since anyone played with the cat, she's bored.
	return MoodImpatient
}

func (c Cat) noise(seed []byte, period time.Duration) behavior.TemporalNoise {
	return behavior.SmoothNoise{
		Underlying: behavior.Md5Noise{Seed: seed},
		Period:     period,
	}
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

// Pat records a new pat for the given Cat.
func Pat(tx *gorm.DB, id CatID) error {
	result := tx.Model(Cat{ID: id}).Updates(map[string]any{
		"pats":       gorm.Expr("pats + 1"),
		"latest_pat": chrono.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("updated %d rows: %w", result.RowsAffected, gorm.ErrRecordNotFound)
	}
	return nil
}
