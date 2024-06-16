package web

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nevkontakte/pat/db"
	"gorm.io/gorm"
)

// Web implements the HTTP server for the pat junkie.
type Web struct {
	StaticFS fs.FS
	DB       *gorm.DB
}

// Bind HTTP handlers to the Echo server.
func (w *Web) Bind(e *echo.Echo) {
	e.GET("/", w.index)
	e.GET("/pat/", w.pat)

	e.Group("/static", middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "",
		Filesystem: http.FS(w.StaticFS),
		// Browse:     true,
	}))
}

// index page handler.
func (w *Web) index(c echo.Context) error {
	splotch, err := db.CatByID(w.DB, db.SplotchID)
	if err != nil { // Should never happen.
		return fmt.Errorf("oops, Splotch went missing 🙀: %w", err)
	}
	data := struct {
		Cat db.Cat
	}{
		Cat: splotch,
	}
	return c.Render(http.StatusOK, "index.html", data)
}

// pat action handler.
func (w *Web) pat(c echo.Context) error {
	if err := db.Pat(w.DB, db.SplotchID); err != nil {
		return fmt.Errorf("failed to pat Splotch: %w", err)
	}
	return c.Redirect(http.StatusFound, "/")
}
