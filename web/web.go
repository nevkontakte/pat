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

	e.Group("/static", middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "",
		Filesystem: http.FS(w.StaticFS),
		// Browse:     true,
	}))
}

func (w *Web) index(c echo.Context) error {
	return c.String(http.StatusOK, "🐈")
}
