package db

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nevkontakte/pat/db/dbtest"
	"gorm.io/gorm"
)

func TestCatByID(t *testing.T) {
	tx := dbtest.InMemory(t)
	tx.AutoMigrate(Cat{})

	c1 := Cat{ID: "black", Name: "Captain Black"}
	c2 := Cat{ID: "red", Name: "Loaf"}
	dbtest.Save(t, tx, &c1, &c2)

	for _, want := range []Cat{c1, c2} {
		t.Run(string(want.ID), func(t *testing.T) {
			got, err := CatByID(tx, want.ID)
			if err != nil {
				t.Errorf("Got: CatByID(%q) returned error: %s. Want: no error.", want.ID, err)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("CatByID(%q) returned diff (-want,+got):\n%s", want.ID, diff)
			}
		})
	}

	t.Run("not found", func(t *testing.T) {
		_, err := CatByID(tx, "stray")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("Got: CatByID(`stray`) returned error: %v. Want: %v.", err, gorm.ErrRecordNotFound)
		}
	})
}

func TestPat(t *testing.T) {
	tx := dbtest.InMemory(t)
	tx.AutoMigrate(Cat{})

	t.Run("exists", func(t *testing.T) {
		now := time.Now()
		c := Cat{
			ID:        "black",
			Name:      "Captain Black",
			Pats:      2,
			LatestPat: now.Add(-time.Minute),
		}
		dbtest.Save(t, tx, c)

		if err := Pat(tx, c.ID); err != nil {
			t.Fatalf("Got: Pat() returned error: %s. Want: no error.", err)
		}
		var got Cat
		dbtest.First(t, tx, &got, "id = ?", c.ID)
		if got.Pats != 3 {
			t.Errorf("Got: recorded pats didn't increment: pats = %v. Want: 3.", got.Pats)
		}
		if got.LatestPat.Before(now) {
			t.Errorf("Got: latest pat time didn't get updated: %v. Want: at or after %v.", got.LatestPat, now)
		}
	})

	t.Run("doesn't exist", func(t *testing.T) {
		if err := Pat(tx, "stray"); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("Got: Pat() returned error: %s. Want: %s.", err, gorm.ErrRecordNotFound)
		}
	})
}
