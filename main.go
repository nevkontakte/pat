package main

import (
	"embed"
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nevkontakte/pat/web"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	bind = flag.String("bind", ":8080", "Address to start the HTTP server at.")
	db   = flag.String("db", "host=localhost user=postgres password=postgres dbname=pat port=5432 sslmode=disable", "Database connection string.")
)

//go:embed static
var staticFS embed.FS

func run(e *echo.Echo) error {
	// Logging
	e.Logger.SetLevel(log.INFO)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db, err := gorm.Open(postgres.Open(*db), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}

	var version string
	db.Raw("SELECT version();").Scan(&version)
	e.Logger.Infof("Connected to %s.", version)

	// Set up HTTP server.
	w := web.Web{
		StaticFS: staticFS,
	}
	w.Bind(e)
	e.Logger.Info(staticFS.ReadDir("static"))

	// Start server
	return e.Start(*bind)
}

func main() {
	flag.Parse()

	e := echo.New()

	if err := run(e); err != nil {
		e.Logger.Fatal(err)
	}
}
