package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type UserController struct {
	userService   app.UserService
	recipeService app.RecipeService
	weekService   app.WeekService
}

func NewUserController(userService app.UserService, recipeService app.RecipeService, weekService app.WeekService) *UserController {
	return &UserController{
		userService:   userService,
		recipeService: recipeService,
		weekService:   weekService,
	}
}

func (uc *UserController) GetUsers(c echo.Context) error {
	users, err := uc.userService.Users()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, users)
}

func (uc *UserController) CreateUser(c echo.Context) error {
	newUser := app.NewUser{}

	var json map[string]interface{}

	if err := c.Bind(&json); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if _, ok := json["name"]; !ok {
		return c.String(http.StatusBadRequest, "JSON body must contain 'name'")
	}
	newUser.Name = json["name"].(string)
	user, err := uc.userService.CreateUser(&newUser)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (uc *UserController) GetUser(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	return c.JSON(http.StatusOK, user)
}

func (uc *UserController) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	err := uc.userService.DeleteUser(id)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}

	return c.NoContent(http.StatusNoContent)
}

func (uc *UserController) GenerateWeek(c echo.Context) error {
	user, err := getUser(uc, c)
	fmt.Println(user)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	currentYear := time.Now().Year()
	_, weekNumber := time.Now().ISOWeek()
	days := app.GenerateDays(currentYear, weekNumber)
	recipes, err := uc.recipeService.Recipes()
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	suggestedRecipes := app.SuggestnRandomRecipes(recipes, len(days))
	for i, day := range days {
		day.ID = uuid.New()
		day.Dinner = suggestedRecipes[i]
	}
	weeks := []*app.Week{
		{
			NewWeek: app.NewWeek{
				Number: weekNumber,
				Days:   days,
			},
			Entity: app.Entity{
				ID: uuid.New(),
			},
		},
	}
	return c.JSON(http.StatusOK, weeks)
}

func (uc *UserController) SaveWeek(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}

	weeks := []*app.NewWeek{}

	if err = c.Bind(&weeks); err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}

	if len(weeks) == 0 {
		httpErr := app.HTTPError{
			Message: "no weeks to save",
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	for _, week := range weeks {
		if week.Number == 0 || week.Number > 52 {
			httpErr := app.HTTPError{
				Message: "number need to be set between 1 and 52",
				Code:    http.StatusUnprocessableEntity,
			}
			return c.JSON(http.StatusUnprocessableEntity, httpErr)
		}
	}

	dbWeeks, err := uc.weekService.CreateWeeks(weeks, user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}

	return c.JSON(http.StatusOK, dbWeeks)
}

func getUser(uc *UserController, c echo.Context) (*app.User, error) {
	fmt.Printf("%+v\n", c)
	id := c.Param("id")
	user, err := uc.userService.User(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user, got: %s", err.Error())
	}
	return user, nil
}
