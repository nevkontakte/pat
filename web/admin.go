package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/nevkontakte/pat/db"
	"github.com/nevkontakte/pat/web/cookie"
	"golang.org/x/crypto/bcrypt"
)

const adminCookieName = "admin_session"

type AdminCookie struct {
	IsAdmin bool
}

// requireAdmin is an Echo middleware that enforces admin authentication.
// Requests without a valid session cookie are redirected to the login page.
func (w *Web) requireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		raw, err := c.Cookie(adminCookieName)
		if err != nil {
			return c.Redirect(http.StatusFound, "/admin/login")
		}
		ac, err := cookie.ParseCookie[AdminCookie](raw.Value, w.Secret)
		if err != nil || !ac.IsAdmin {
			return c.Redirect(http.StatusFound, "/admin/login")
		}
		return next(c)
	}
}

type loginData struct{ Error error }

func (w *Web) adminLogin(c *echo.Context) error {
	return c.Render(http.StatusOK, "login.html", &loginData{})
}

func (w *Web) adminLoginPost(c *echo.Context) error {
	password := c.FormValue("password")
	if err := bcrypt.CompareHashAndPassword(w.AdminPasswordHash, []byte(password)); err != nil {
		return c.Render(http.StatusOK, "login.html", &loginData{
			Error: fmt.Errorf("Wrong password."),
		})
	}
	value, err := cookie.SaveCookie(AdminCookie{IsAdmin: true}, w.Secret)
	if err != nil {
		return fmt.Errorf("failed to create admin session cookie: %w", err)
	}
	c.SetCookie(&http.Cookie{
		Name:     adminCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Scheme() == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60,
	})
	return c.Redirect(http.StatusFound, "/admin/")
}

func (w *Web) adminLogout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:   adminCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	return c.Redirect(http.StatusFound, "/")
}

func (w *Web) adminDashboard(c *echo.Context) error {
	splotch, err := db.CatByID(w.DB, db.SplotchID)
	if err != nil {
		return fmt.Errorf("failed to load cat: %w", err)
	}

	var lastVisit db.Journal
	lastVisitTime := time.Time{}
	if result := w.DB.Where("cat_id = ? AND type = ?", db.SplotchID, db.EventVisit).
		Order("created_at desc").First(&lastVisit); result.Error == nil {
		lastVisitTime = lastVisit.CreatedAt
	}

	data := struct {
		Mood      db.Mood
		LastVisit time.Time
		LastPat   time.Time
	}{
		Mood:      splotch.Mood(),
		LastVisit: lastVisitTime,
		LastPat:   splotch.LatestPat,
	}
	return c.Render(http.StatusOK, "admin.html", data)
}
