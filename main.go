package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"nrdev.se/mealshuffler/app"
	"nrdev.se/mealshuffler/sqlite"
)

type server struct {
	userService   app.UserService
	weekService   app.WeekService
	recipeService app.RecipeService
}

func main() {

	e := echo.New()

	db, err := sqlite.NewDB()
	if err != nil {
		e.Logger.Fatal(err)
	}

	recipeService := sqlite.NewRecipeService(db)
	userService := sqlite.NewUserService(db)
	weekService := sqlite.NewWeekService(db)

	setupMiddleware(e)

	srv := server{
		userService:   userService,
		weekService:   weekService,
		recipeService: recipeService,
	}

	e.Static("/assets", "public/assets")

	srv.setupRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
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
			Code:    http.StatusBadRequest,
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
		return next(c)
	}
}
