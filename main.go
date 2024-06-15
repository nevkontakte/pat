package main

import (
	"flag"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	bind = flag.String("bind", ":8080", "Address to start the HTTP server at.")
	db   = flag.String("db", "host=localhost user=postgres password=postgres dbname=pat port=5432 sslmode=disable", "Database connection string.")
)

func main() {
	// Echo instance
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)

	db, err := gorm.Open(postgres.Open(*db), &gorm.Config{})
	if err != nil {
		e.Logger.Fatalf("Failed to connect to the database at %q: %v", err)
	}
	var version string
	db.Raw("SELECT version();").Scan(&version)
	e.Logger.Infof("Connected to %s.", version)

	// Start server
	e.Logger.Fatal(e.Start(*bind))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "🐈")
}
