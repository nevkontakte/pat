package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Web implements the HTTP server for the pat junkie.
type Web struct{}

// Bind HTTP handlers to the Echo server.
func (w *Web) Bind(e *echo.Echo) {
	e.GET("/", w.index)
}

func (w *Web) index(c echo.Context) error {
	return c.String(http.StatusOK, "🐈")
}
