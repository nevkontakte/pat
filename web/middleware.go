package web

import (
	"github.com/labstack/echo/v4"
	"github.com/nevkontakte/pat/db"
)

const visitorKey = "visitor"

// VisitorMiddleware populates the Echo context with the current visitor.
func VisitorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set(visitorKey, db.CurrentVisitor(c))
		return next(c)
	}
}

// VisitorFromContext retrieves the current visitor from the Echo context.
func VisitorFromContext(c echo.Context) *db.Visitor {
	if v, ok := c.Get(visitorKey).(db.Visitor); ok {
		return &v
	}
	return &db.Visitor{}
}
