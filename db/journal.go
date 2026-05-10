package db

import (
	"net/netip"
	"time"

	"github.com/labstack/echo/v4"
)

// Visitor represents information about a visitor.
//
// It doesn't uniquely identify the visitor, unless they volunteerely "introduce" themselves
// (something we may implement in future).
type Visitor struct {
	// Visitor's IP address.
	Addr Addr
	// Visitor's User Agent.
	Agent string
	// Referrer is the HTTP Referer header value.
	Referrer string
}

// CurrentVisitor populates the Visitor instance from the request.
//
// Visitor identification is best-effort, if some aspects can't be identified,
// we just leave them empty.
func CurrentVisitor(c echo.Context) Visitor {
	v := Visitor{
		Agent:    c.Request().UserAgent(),
		Referrer: c.Request().Referer(),
	}
	if addr, err := netip.ParseAddr(c.RealIP()); err == nil {
		v.Addr = Addr(addr)
	}
	return v
}

// Game world event type.
type EventType uint16

const (
	EventUnknown EventType = iota // Unknown, default value. Should never happen.
	EventVisit                    // The cat was visited without explicit interaction.
	EventPat                      // The cat received a pat.
)

// Event describes a game world event.
type Event struct {
	Type        EventType
	Description string // Human-readable description.
}

// Journal records a notable event that happened in the game world, which could be autonomous events or user interactions.
type Journal struct {
	ID uint64 `gorm:"primaryKey"`
	// CreatedAt contains journal record creation time. Auto-populated by Gorm.
	CreatedAt time.Time

	// Visitor that triggered event. Nil if no visitor was involved.
	Visitor *Visitor `gorm:"embedded"`

	// CatID identifies which cat the event happened to, most of the time this would be Splotch.
	CatID CatID
	Cat   Cat

	// Event metadata that the journal record represents.
	Event Event `gorm:"embedded"`
}
