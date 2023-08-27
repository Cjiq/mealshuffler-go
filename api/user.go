package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type UserController struct {
	userService   app.UserService
	recipeService app.RecipeService
}

func NewUserController(userService app.UserService, recipeService app.RecipeService) *UserController {
	return &UserController{
		userService:   userService,
		recipeService: recipeService,
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
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusBadRequest, "ID must be an integer")
	}
	err = uc.userService.DeleteUser(idNum)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error")
	}

	return c.NoContent(http.StatusNoContent)
}

func (uc *UserController) GenerateWeek(c echo.Context) error {
	user, err := getUser(uc, c)
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
	user.Weeks = []*app.Week{
		{
			Number: weekNumber,
			Entity: app.Entity{
				ID: uuid.New(),
			},
			Days: days,
		},
	}
	return c.JSON(http.StatusOK, user)
}

func getUser(uc *UserController, c echo.Context) (*app.User, error) {
	id := c.Param("id")
	user, err := uc.userService.User(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user, got: %s", err.Error())
	}
	return user, nil
}
