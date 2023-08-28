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
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.Use(checkRequestContentTypeJSON)

	db, err := sqlite.NewDB()
	if err != nil {
		e.Logger.Fatal(err)
	}

	recipeService := sqlite.NewRecipeService(db)
	recipeController := api.NewRecipeController(recipeService)

	userService := sqlite.NewUserService(db)
	userController := api.NewUserController(userService, recipeService)

	e.GET("/", hello)
	e.GET("/api/users", userController.GetUsers)
	e.POST("/api/users", userController.CreateUser)
	e.GET("/api/users/:id", userController.GetUser)
	e.DELETE("/api/users/:id", userController.DeleteUser)
	e.GET("/api/users/:id/generate", userController.GenerateWeek)
	e.POST("/api/users/:id/weeks", userController.SaveWeek)

	e.GET("/api/recipes", recipeController.GetRecipes)
	e.POST("/api/recipes", recipeController.CreateRecipe)

	e.Logger.Fatal(e.Start(":8080"))
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, world!")
}

func checkRequestContentTypeJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == http.MethodGet {
			return next(c)
		}
		if c.Request().Header.Get("Content-Type") != "application/json" {
			return c.String(http.StatusBadRequest, "Content-Type must be application/json")
		}
		return next(c)
	}
}
