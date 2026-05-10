package db

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
	"github.com/nevkontakte/pat/db/dbtest"
)

func TestJournal_SaveAndRestore(t *testing.T) {
	tx := dbtest.InMemory(t)
	tx.AutoMigrate(&Cat{}, &Journal{})

	now := time.Now()

	j := Journal{
		Visitor: &Visitor{
			Addr:     Addr(netip.MustParseAddr("127.0.0.1")),
			Agent:    "TestAgent/1.0",
			Referrer: "https://example.com/",
		},
		CatID: CatID("black"),
		Cat: Cat{
			ID:        CatID("black"),
			Name:      "Captain Black",
			Pats:      2,
			LatestPat: now.Add(-time.Minute),
		},
		Event: Event{
			Type:        EventPat,
			Description: "A friendly pat",
		},
	}

	dbtest.Save(t, tx, &j)

	var got Journal
	dbtest.First(t, tx.Preload("Cat"), &got, "id = ?", j.ID)

	if diff := cmp.Diff(j, got, cmpopts.EquateComparable(Addr{})); diff != "" {
		t.Errorf("Journal mismatch (-want +got):\n%s", diff)
	}
}

func TestCurrentVisitor(t *testing.T) {
	t.Run("populates IP, user agent, and referrer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		req.Header.Set("Referer", "https://example.com/")

		c := echo.New().NewContext(req, httptest.NewRecorder())

		got := CurrentVisitor(c)
		want := Visitor{
			Addr:     Addr(netip.MustParseAddr("127.0.0.1")),
			Agent:    "TestAgent/1.0",
			Referrer: "https://example.com/",
		}

		if diff := cmp.Diff(want, got, cmpopts.EquateComparable(Addr{})); diff != "" {
			t.Errorf("CurrentVisitor() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid IP leaves zero value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "not-an-ip"
		req.Header.Set("User-Agent", "")

		c := echo.New().NewContext(req, httptest.NewRecorder())

		got := CurrentVisitor(c)
		want := Visitor{}

		if diff := cmp.Diff(want, got, cmpopts.EquateComparable(Addr{})); diff != "" {
			t.Errorf("CurrentVisitor() mismatch (-want +got):\n%s", diff)
		}
	})
}
