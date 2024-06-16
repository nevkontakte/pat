package main

import (
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nevkontakte/pat/db"
	"github.com/nevkontakte/pat/static"
	"github.com/nevkontakte/pat/tmpl"
	"github.com/nevkontakte/pat/web"
)

var (
	bind = flag.String("bind", ":8080", "Address to start the HTTP server at.")
	dsn  = flag.String("db", "host=localhost user=postgres password=postgres dbname=pat port=5432 sslmode=disable", "Database connection string.")
)

func run(e *echo.Echo) error {
	// Logging
	e.Logger.SetLevel(log.INFO)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	dbconn, err := db.Postgres(*dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	if err := db.Bootstrap(dbconn); err != nil {
		return fmt.Errorf("failed to bootstrap the database: %w", err)
	}

	e.Renderer, err = tmpl.Load()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Set up HTTP server.
	w := web.Web{
		StaticFS: static.StaticFS,
		DB:       dbconn,
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
