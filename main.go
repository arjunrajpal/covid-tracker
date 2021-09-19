package main

import (
	_ "app/docs"
	"app/resource"
	"github.com/labstack/echo/v4"
	"github.com/swaggo/echo-swagger"
	"net/http"
	"os"
)

// @title Covid active cases tracker
// @version 1.0
// @description Swagger API for Golang Project Covid active cases tracker.

// @contact.name Arjun Rajpal
// @contact.email rajpal.arjun@yahoo.cin

// @BasePath /
func main() {

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/api/v1/getActiveCases", func(c echo.Context) error { return resource.GetActiveCases(c) })

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
