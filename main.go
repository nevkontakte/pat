package main

import (
	"flag"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	bind = flag.String("bind", ":8080", "Address to start the HTTP server at.")
	db   = flag.String("db", "", "Database connection string.")
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)

	// Start server
	e.Logger.Fatal(e.Start(*bind))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "🐈")
}
