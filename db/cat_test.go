package db

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nevkontakte/pat/db/dbtest"
	"gorm.io/gorm"
)

func TestCatByID(t *testing.T) {
	tx := dbtest.InMemory(t)
	tx.AutoMigrate(&Cat{})

	c1 := Cat{ID: "black", Name: "Captain Black"}
	c2 := Cat{ID: "red", Name: "Loaf"}
	dbtest.Save(t, tx, &c1, &c2)

	for _, want := range []Cat{c1, c2} {
		t.Run(want.ID, func(t *testing.T) {
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
