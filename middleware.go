package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"nrdev.se/mealshuffler/app"
)

func setupMiddleware(e *echo.Echo) {
	// CORS origin as a command line flag
	corsOrigin := flag.String("cors-origin", "*", "CORS origin")
	flag.Parse()

	corsConfig := middleware.CORSConfig{
		AllowOrigins: []string{*corsOrigin},
	}

	log.Printf("CORS config: %+v", corsConfig)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORSWithConfig(corsConfig))
	e.Use(checkRequestContentTypeJSON)
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic in HTTPErrorHandler: ", r)
				httpErr := app.HTTPError{
					Message: "internal server error",
					Code:    http.StatusInternalServerError,
				}
				c.JSON(http.StatusInternalServerError, httpErr)
				panic(r)
			}
		}()

		resErr, ok := err.(app.HTTPError)
		if ok {
			c.JSON(resErr.Code, resErr)
			return
		}

		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}

		c.JSON(httpErr.Code, httpErr)
	}
}
