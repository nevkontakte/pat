package db

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nevkontakte/pat/chrono/chronotest"
	"github.com/nevkontakte/pat/db/dbtest"
	"gorm.io/gorm"
)

func TestCatID(t *testing.T) {
	t.Run("Seed", func(t *testing.T) {
		// A single, representative test case for Seed.
		id := CatID("mycat")
		cue := "saltycue"
		want := []byte("saltycuemycat")
		got := id.Seed(cue)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("CatID(%q).Seed(%q) mismatch (-want +got):\n%s", id, cue, diff)
		}
	})

	t.Run("Name", func(t *testing.T) {
		testCases := []struct {
			name string
			id   CatID
			want string
		}{
			{"simple lowercase", CatID("mycat"), "Mycat"},
			{"with hyphen", CatID("my-cat"), "My-Cat"},
			{"already partially title cased", CatID("MyCat"), "Mycat"},
			{"splotch (constant example)", SplotchID, "Splotch"},
			{"empty string", CatID(""), ""},
			{"single character", CatID("a"), "A"},
			{"all uppercase", CatID("UPPERCAT"), "Uppercat"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if got := tc.id.Name(); got != tc.want {
					t.Errorf("CatID(%q).Name() got %q, want %q", tc.id, got, tc.want)
				}
			})
		}
	})
}

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

func TestCat_Mood(t *testing.T) {
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	chronotest.OverrideNow(t, now)

	// generalTestCat is used for most mood calculations, its ID affects noise generation.
	generalTestCat := Cat{ID: "testcat", Name: "Test Cat Moods"}

	testCases := []struct {
		name         string
		cat          Cat       // Cat struct, ID is important for noise
		latestPat    time.Time // This will be set on the cat
		expectedMood Mood
	}{
		{
			name:         "MoodPat - petted just now",
			cat:          generalTestCat,
			latestPat:    now,
			expectedMood: MoodPat,
		},
		{
			name:         "MoodPat - petted 4.9s ago",
			cat:          generalTestCat,
			latestPat:    now.Add(-4*time.Second - 900*time.Millisecond),
			expectedMood: MoodPat,
		},
		{
			name:         "MoodIdleHappy - petted 5s ago (just after MoodPat window)",
			cat:          generalTestCat,
			latestPat:    now.Add(-5 * time.Second),
			expectedMood: MoodIdleHappy,
		},
		{
			name:         "MoodIdleHappy - petted 29m59s ago",
			cat:          generalTestCat,
			latestPat:    now.Add(-29*time.Minute - 59*time.Second),
			expectedMood: MoodIdleHappy,
		},
		{
			name:         "MoodIdle - petted 1 day ago",
			cat:          generalTestCat,
			latestPat:    now.Add(-1 * 24 * time.Hour),
			expectedMood: MoodIdle,
		},
		{
			name:         "MoodIdle - petted just under 7 days ago",
			cat:          generalTestCat,
			latestPat:    now.Add(-7*24*time.Hour + 1*time.Minute),
			expectedMood: MoodIdle,
		},
		{
			name:         "MoodImpatient - petted 7 days ago",
			cat:          generalTestCat,
			latestPat:    now.Add(-7 * 24 * time.Hour),
			expectedMood: MoodImpatient,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			catToTest := tc.cat
			catToTest.LatestPat = tc.latestPat
			mood := catToTest.Mood()
			if mood != tc.expectedMood {
				t.Errorf("For cat ID '%s', LatestPat %v (sincePat %v), expected mood %s, but got %s.",
					tc.cat.ID, tc.latestPat, now.Sub(tc.latestPat), tc.expectedMood, mood)
			}
		})
	}

	t.Run("MoodSwingCausesBothHappyAndIdleOverTimeRange", func(t *testing.T) {
		cat := Cat{
			ID:        SplotchID,
			Name:      "Splotch",
			LatestPat: now,
		}
		moods := map[Mood]int{}
		latest := Mood("")
		for delay := 30 * time.Minute; delay <= 3*time.Hour+30*time.Minute; delay += time.Minute {
			var current Mood
			chronotest.OverrideScope(now.Add(delay), func() {
				current = cat.Mood()
			})

			if current != latest {
				moods[current]++
				latest = current
			}
		}

		want := map[Mood]int{
			MoodIdleHappy: 6,
			MoodIdle:      6,
		}
		if diff := cmp.Diff(want, moods); diff != "" {
			t.Errorf("Mood swing test for SplotchID mismatch (-want +got):\n%s", diff)
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
