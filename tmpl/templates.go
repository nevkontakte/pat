package tmpl

import (
	"embed"
	"fmt"
	"io"

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

// Load and parse all embedded templates.
func Load() (*Template, error) {
	t, err := template.ParseFS(template.TrustedFSFromEmbed(templateFS), "*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded templates: %w", err)
	}
	return &Template{templates: t}, nil
}
