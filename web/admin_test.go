package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nevkontakte/pat/db"
	"github.com/nevkontakte/pat/db/dbtest"
	"github.com/nevkontakte/pat/tmpl"
	"github.com/nevkontakte/pat/web/cookie"
	"golang.org/x/crypto/bcrypt"
)

func newTestWeb(t *testing.T) *Web {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	dbconn := dbtest.InMemory(t)
	if err := db.Bootstrap(dbconn); err != nil {
		t.Fatalf("db.Bootstrap: %v", err)
	}
	return &Web{
		Secret:            []byte("testsecret"),
		AdminPasswordHash: hash,
		DB:                dbconn,
	}
}

func newTestEcho(t *testing.T) *echo.Echo {
	t.Helper()
	e := echo.New()
	renderer, err := tmpl.Load()
	if err != nil {
		t.Fatalf("tmpl.Load: %v", err)
	}
	e.Renderer = renderer
	return e
}

func validAdminCookieValue(t *testing.T, w *Web) string {
	t.Helper()
	value, err := cookie.SaveCookie(AdminCookie{IsAdmin: true}, w.Secret)
	if err != nil {
		t.Fatalf("cookie.SaveCookie: %v", err)
	}
	return string(value)
}

// requireAdmin middleware tests

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name           string
		cookie         func(w *Web) *http.Cookie
		wantNextCalled bool
		wantCode       int
		wantLocation   string
	}{
		{
			name:           "no cookie",
			wantNextCalled: false,
			wantCode:       http.StatusFound,
			wantLocation:   "/admin/login",
		},
		{
			name: "invalid cookie",
			cookie: func(w *Web) *http.Cookie {
				return &http.Cookie{Name: adminCookieName, Value: "deadbeef"}
			},
			wantNextCalled: false,
			wantCode:       http.StatusFound,
			wantLocation:   "/admin/login",
		},
		{
			name: "valid cookie",
			cookie: func(w *Web) *http.Cookie {
				return &http.Cookie{Name: adminCookieName, Value: validAdminCookieValue(t, w)}
			},
			wantNextCalled: true,
			wantCode:       http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newTestWeb(t)
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie(w))
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			called := false
			err := w.requireAdmin(func(c echo.Context) error {
				called = true
				return c.String(http.StatusOK, "ok")
			})(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if called != tc.wantNextCalled {
				t.Errorf("next called = %v, want %v", called, tc.wantNextCalled)
			}
			if rec.Code != tc.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantCode)
			}
			if tc.wantLocation != "" {
				if loc := rec.Header().Get("Location"); loc != tc.wantLocation {
					t.Errorf("Location = %q, want %q", loc, tc.wantLocation)
				}
			}
		})
	}
}

// Login handler tests

func TestAdminLogin_RendersForm(t *testing.T) {
	w := newTestWeb(t)
	e := newTestEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := w.adminLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `type="password"`) {
		t.Error("response should contain a password input")
	}
	if strings.Contains(rec.Body.String(), "Wrong password") {
		t.Error("error message should not appear on initial GET")
	}
}

// Login POST handler tests

func TestAdminLoginPost_CorrectPassword(t *testing.T) {
	w := newTestWeb(t)
	e := echo.New()

	form := url.Values{"password": {"testpass"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := w.adminLoginPost(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/admin/" {
		t.Errorf("expected redirect to /admin/, got %q", loc)
	}

	var adminCookie *http.Cookie
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == adminCookieName {
			adminCookie = ck
			break
		}
	}
	if adminCookie == nil {
		t.Fatal("admin_session cookie should be set after successful login")
	}
	ac, err := cookie.ParseCookie[AdminCookie](adminCookie.Value, w.Secret)
	if err != nil {
		t.Fatalf("cookie.ParseCookie failed: %v", err)
	}
	if !ac.IsAdmin {
		t.Error("cookie payload should have IsAdmin=true")
	}
	if !adminCookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if adminCookie.MaxAge <= 0 {
		t.Errorf("cookie should have positive MaxAge, got %d", adminCookie.MaxAge)
	}
}

func TestAdminLoginPost_WrongPassword(t *testing.T) {
	w := newTestWeb(t)
	e := newTestEcho(t)

	form := url.Values{"password": {"wrongpass"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := w.adminLoginPost(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Wrong password") {
		t.Error("response should contain the error message")
	}
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == adminCookieName {
			t.Error("admin_session cookie should not be set after failed login")
		}
	}
}

// Logout handler tests

func TestAdminLogout(t *testing.T) {
	w := newTestWeb(t)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/admin/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := w.adminLogout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("expected redirect to /, got %q", loc)
	}

	cleared := false
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == adminCookieName {
			cleared = true
			if ck.MaxAge != -1 {
				t.Errorf("cookie MaxAge should be -1 to delete it, got %d", ck.MaxAge)
			}
		}
	}
	if !cleared {
		t.Error("logout should set the admin_session cookie with MaxAge=-1")
	}
}

// Dashboard handler tests

func TestAdminDashboard(t *testing.T) {
	w := newTestWeb(t)
	e := newTestEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := w.adminDashboard(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Splotch") {
		t.Error("dashboard should show the cat name")
	}
	for _, label := range []string{"Mood", "Last visit", "Last pat"} {
		if !strings.Contains(body, label) {
			t.Errorf("dashboard should show %q label", label)
		}
	}
	if !strings.Contains(body, `href="/"`) {
		t.Error("dashboard should have a home link")
	}
	if !strings.Contains(body, `href="/admin/logout"`) {
		t.Error("dashboard should have a logout link")
	}
}
