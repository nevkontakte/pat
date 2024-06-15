package main

import (
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nevkontakte/pat/static"
	"github.com/nevkontakte/pat/tmpl"
	"github.com/nevkontakte/pat/web"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	bind = flag.String("bind", ":8080", "Address to start the HTTP server at.")
	db   = flag.String("db", "host=localhost user=postgres password=postgres dbname=pat port=5432 sslmode=disable", "Database connection string.")
)

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

	e.Renderer, err = tmpl.Load()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Set up HTTP server.
	w := web.Web{
		StaticFS: static.StaticFS,
	}
	w.Bind(e)

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
