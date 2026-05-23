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
	StaticFS          fs.FS
	DB                *gorm.DB
	AdminPasswordHash []byte // Bcrypt hash of the admin password. Admin routes are disabled if empty.
	Secret            []byte // Server-side signing secret for session cookies. Admin routes are disabled if empty.
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

	if len(w.AdminPasswordHash) > 0 && len(w.Secret) > 0 {
		e.GET("/admin/login", w.adminLogin)
		e.POST("/admin/login", w.adminLoginPost)

		admin := e.Group("/admin", w.requireAdmin)
		admin.GET("/", w.adminDashboard)
		admin.GET("/logout", w.adminLogout)
	}
}

// index page handler.
func (w *Web) index(c echo.Context) error {
	splotch, err := db.CatByID(w.DB, db.SplotchID)
	if err != nil { // Should never happen.
		return fmt.Errorf("oops, Splotch went missing 🙀: %w", err)
	}
	if err := w.recordJournal(c, db.Event{Type: db.EventVisit}); err != nil {
		return err
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

	if err := w.recordJournal(c, db.Event{Type: db.EventPat}); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func (w *Web) recordJournal(c echo.Context, e db.Event) error {
	result := w.DB.Save(&db.Journal{
		Visitor: VisitorFromContext(c),
		CatID:   db.SplotchID,
		Event:   e,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to add a journal entry: %s", result.Error)
	}
	return nil
}
