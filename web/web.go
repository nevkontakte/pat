package web

import (
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Web implements the HTTP server for the pat junkie.
type Web struct {
	StaticFS fs.FS
}

// Bind HTTP handlers to the Echo server.
func (w *Web) Bind(e *echo.Echo) {
	e.GET("/", w.index)

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "",
		Browse:     false,
		Filesystem: http.FS(w.StaticFS),
	}))
}

func (w *Web) index(c echo.Context) error {
	return c.String(http.StatusOK, "🐈")
}
