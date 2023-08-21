package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"nrdev.se/mealshuffler/api"
	"nrdev.se/mealshuffler/sqlite"
)

func main() {

	fmt.Println("Hello World")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db, err := sqlite.NewDB()
	if err != nil {
		e.Logger.Fatal(err)
	}

	userService := sqlite.NewUserService(db)
	userController := api.NewUserController(userService)

	e.GET("/", hello)
	e.GET("/api/users", userController.GetUsers)
	e.POST("/api/users", userController.CreateUser)
	e.GET("/api/users/:id", userController.GetUser)
	e.DELETE("/api/users/:id", userController.DeleteUser)

	e.Logger.Fatal(e.Start(":8080"))
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, world!")
}
