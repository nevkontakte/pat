package db

import (
	"testing"

	"github.com/nevkontakte/pat/db/dbtest"
)

func TestBootstrap(t *testing.T) {
	dbconn := dbtest.InMemory(t)

	// First bootstrap should set the database up.
	if err := Bootstrap(dbconn); err != nil {
		t.Fatalf("Got: First Bootstrap() returned error: %s. Want: no error.", err)
	}

	// First bootstrap should create a record for Splotch with a single pat.
	var splotch Cat
	dbtest.First(t, dbconn, &splotch, "id = ?", SplotchID)
	if splotch.Pats != 1 {
		t.Fatalf("Got: splotch.Pats = %v. Want: 1", splotch.Pats)
	}

	// Second Bootstrap should not change anything in the database.
	splotch.Pats = 2
	dbtest.Save(t, dbconn, &splotch)
	if err := Bootstrap(dbconn); err != nil {
		t.Fatalf("Got: Second Bootstrap() returned error: %s. Want: no error.", err)
	}
	var splotch2 Cat
	dbtest.First(t, dbconn, &splotch2, "id = ?", SplotchID)
	if splotch2.Pats != 2 {
		t.Fatalf("Got: Second Bootstrap() changed number of pats to %v. Want: 2", splotch2.Pats)
	}
}
