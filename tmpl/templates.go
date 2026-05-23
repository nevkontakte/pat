package tmpl

import (
	"embed"
	"fmt"
	"io"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/safehtml/template"
	"github.com/labstack/echo/v4"
)

//go:embed *.html
var templateFS embed.FS

// Template implements echo.Renderer interface for safehtml/template
type Template struct {
	templates *template.Template
}

// Render requested template with the provided context.
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// since formats a time as a human-readable relative duration (e.g. "3 hours ago").
// Returns "never" for the zero time.
func since(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return humanize.Time(t)
}

// Load and parse all embedded templates.
func Load() (*Template, error) {
	t, err := template.New("").
		Funcs(template.FuncMap{"since": since}).
		ParseFS(template.TrustedFSFromEmbed(templateFS), "*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded templates: %w", err)
	}
	return &Template{templates: t}, nil
}
