package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"

	"nrdev.se/mealshuffler/api"
	"nrdev.se/mealshuffler/app"
	"nrdev.se/mealshuffler/sqlite"
)

type server struct {
	userService app.UserService
}

func main() {

	// CORS origin as a command line flag

	corsOrigin := flag.String("cors-origin", "*", "CORS origin")
	flag.Parse()

	e := echo.New()

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
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    err.(*echo.HTTPError).Code,
		}
		c.JSON(http.StatusInternalServerError, httpErr)
	}

	db, err := sqlite.NewDB()
	if err != nil {
		e.Logger.Fatal(err)
	}

	recipeService := sqlite.NewRecipeService(db)
	recipeController := api.NewRecipeController(recipeService)

	userService := sqlite.NewUserService(db)
	weekService := sqlite.NewWeekService(db)
	userController := api.NewUserController(userService, recipeService, weekService)

	weekController := api.NewWeekController(weekService)

	srv := server{
		userService: userService,
	}
	e.POST("/login", srv.login)

	api := e.Group("/api")
	api.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(key string, _ echo.Context) (bool, error) {
			_, err := userService.ValidateUserToken(key)
			if err != nil {
				log.Println("error validating token: ", err)
				return false, err
			}
			return true, nil
		},
		ErrorHandler: func(err error, _ echo.Context) error {
			httpErr := app.HTTPError{
				Message: err.Error(),
				Code:    http.StatusUnauthorized,
			}
			return echo.NewHTTPError(http.StatusUnauthorized, httpErr)
		},
	}))

	api.GET("/ping", ping)

	api.GET("/users/:id/generate", userController.GenerateWeek)
	api.POST("/users/:id/weeks", userController.SaveWeek)
	api.DELETE("/users/:id/weeks/:weekID", userController.DeleteWeek)
	api.GET("/users/:id/weeks/next", userController.NextWeekNumber)
	api.POST("/users/:id/weeks/:weekID/suggest", userController.GenerateRecipeAlternative)
	api.PUT("/users/:id/weeks/:weekID", userController.UpdateWeek)
	api.PUT("/users/:id/weeks", userController.UpdateWeeks)
	api.POST("/users/:id/recipes", recipeController.CreateRecipe)

	api.GET("/recipes", recipeController.GetRecipes)
	api.DELETE("/recipes", recipeController.DeleteRecipes)
	api.PUT("/recipes", recipeController.UpdateRecipe)

	api.GET("/users/:id/weeks/:year", weekController.GetWeeks)
	api.GET("/users/:id/weeks/last", weekController.GetLastGeneratedWeek)
	api.DELETE("/users/:id/weeks/:year/all", weekController.DeleteWeeks)

	admin := e.Group("/api/admin")
	admin.Use(srv.AdminMiddleware)
	admin.GET("/ping", ping)
	admin.POST("/users", userController.CreateUser)
	admin.GET("/users", userController.GetUsers)
	admin.GET("/users/:id", userController.GetUser)
	admin.DELETE("/users/:id", userController.DeleteUser)

	e.Logger.Fatal(e.Start(":8080"))
}

func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func (s *server) login(c echo.Context) error {
	type user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var u user
	if err := c.Bind(&u); err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}

	userHash, err := s.userService.GetUserHash(u.Username)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	err = bcrypt.CompareHashAndPassword(userHash, []byte(u.Password))
	if err != nil {
		httpErr := app.HTTPError{
			Message: "invalid username or password",
			Code:    http.StatusUnauthorized,
		}
		return c.JSON(http.StatusUnauthorized, httpErr)
	}
	dbUser, err := s.userService.UserByUserName(u.Username)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}

	newBearerToken, err := generateBearerToken(48)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	err = s.userService.SaveUserToken(dbUser.ID.String(), newBearerToken)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token":      newBearerToken,
		"token_type": "bearer",
		"user":       dbUser,
	})
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

func generateBearerToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	// Calculate the number of bytes needed
	byteLength := (length * 6) / 8 // 6 bits per Base64 character

	// Generate random bytes
	randomBytes := make([]byte, byteLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes to Base64
	token := base64.URLEncoding.EncodeToString(randomBytes)

	// Trim any padding characters (=) from the token
	token = strings.TrimRight(token, "=")

	return token, nil
}
func (s *server) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		adminToken, err := s.userService.GetUserToken("cjiq")
		fmt.Printf("%s == %s", token, adminToken)
		if err != nil {
			httpErr := app.HTTPError{
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			}
			return c.JSON(http.StatusInternalServerError, httpErr)
		}

		if token != adminToken {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}
		fmt.Println("==== ADMIN ACCESS ====")
		return next(c)
	}
}
