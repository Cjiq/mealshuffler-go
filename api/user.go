package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

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
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	return c.JSON(http.StatusOK, users)
}

func (uc *UserController) CreateUser(c echo.Context) error {
	var newUser app.NewUser

	if err := c.Bind(&newUser); err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	if newUser.Name == "" {
		httpErr := app.HTTPError{
			Message: "name is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	if newUser.Username == "" {
		httpErr := app.HTTPError{
			Message: "username is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	if newUser.Password == "" {
		httpErr := app.HTTPError{
			Message: "password is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 14)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "failed to hash password",
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	user, err := uc.userService.CreateUser(&newUser, newHash)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
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
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	weekNumber := -1
	now := time.Now()
	currentYear, backupWeek := now.ISOWeek()
	weekNumber, err = uc.weekService.NextWeekNumber(user.ID.String())
	if err != nil {
		if strings.Contains(err.Error(), "converting NULL to int is unsupported") {
			weekNumber = backupWeek
		} else {
			httpErr := app.HTTPError{
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			}

			return c.JSON(http.StatusInternalServerError, httpErr)
		}
	}
	if weekNumber == 0 {
		weekNumber = backupWeek
	}

	previousWeeks, err := uc.weekService.Weeks(user.ID.String(), currentYear)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	var prevDays []*app.Day
	for _, week := range previousWeeks {
		if week.Number == weekNumber-1 {
			prevDays = week.Days
		}
	}

	recipes, err := uc.recipeService.UserRecipes(user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	if len(recipes) == 0 {
		httpErr := app.HTTPError{
			Message: "no recipes found to generate from",
			Code:    http.StatusNotFound,
		}
		return c.JSON(httpErr.Code, httpErr)
	}

	days := app.GenerateDays(currentYear, weekNumber)
	for _, day := range days {
		day.ID = uuid.New()
		day.Dinner = app.PickRecipeForDay(day, append(prevDays, days...), recipes)
	}
	weeks := []*app.Week{
		{
			NewWeek: app.NewWeek{
				Number: weekNumber,
				Days:   days,
				Year:   currentYear,
			},
			Entity: app.Entity{
				ID: uuid.New(),
			},
		},
	}
	newWeek := &app.NewWeek{}
	newWeek.Days = weeks[0].Days
	newWeek.Number = -1
	newWeek.Year = weeks[0].Year

	week := &app.Week{
		NewWeek: *newWeek,
		Entity: app.Entity{
			ID: uuid.New(),
		},
	}

	var lastGeneratedWeek *app.Week
	lastGeneratedWeek, err = uc.weekService.LastGeneratedWeek(user.ID.String())
	if err != nil || lastGeneratedWeek == nil {
		if err.Error() == "week not found" || lastGeneratedWeek == nil {
			lastGeneratedWeek, err = uc.weekService.CreateWeek(newWeek, user.ID.String())
			if err != nil {
				httpErr := app.HTTPError{
					Message: err.Error(),
					Code:    http.StatusInternalServerError,
				}
				return c.JSON(http.StatusInternalServerError, httpErr)
			}
		}
	}
	week.ID = lastGeneratedWeek.ID
	_, err = uc.weekService.UpdateWeek(week, user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
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

	// nextWeekNumber, err := uc.weekService.NextWeekNumber(user.ID.String())
	// if err != nil {
	// 	httpErr := app.HTTPError{
	// 		Message: err.Error(),
	// 		Code:    http.StatusInternalServerError,
	// 	}
	// 	return c.JSON(http.StatusInternalServerError, httpErr)
	// }
	// c.Response().Header().Set("X-Next-Week-Number", fmt.Sprintf("%d", nextWeekNumber))

	return c.JSON(http.StatusCreated, dbWeeks)
}

func (uc *UserController) DeleteWeek(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	weekID := c.Param("weekID")
	err = uc.weekService.DeleteWeek(weekID, user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	nextWeekNumber, err := uc.weekService.NextWeekNumber(user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	if nextWeekNumber == 0 {
		_, nextWeekNumber = time.Now().ISOWeek()
	}
	c.Response().Header().Set("X-Next-Week-Number", fmt.Sprintf("%d", nextWeekNumber))

	return c.NoContent(http.StatusNoContent)
}

func (uc *UserController) NextWeekNumber(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}

	nextWeekNumber, err := uc.weekService.NextWeekNumber(user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}

	return c.JSON(http.StatusOK, nextWeekNumber)
}

func (uc *UserController) GenerateRecipeAlternative(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	weekID := c.Param("weekID")
	var day *app.Day
	if err = c.Bind(&day); err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	var week *app.Week
	isLastGenerated := false
	week, err = uc.weekService.Week(weekID, user.ID.String())
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			week, err = uc.weekService.LastGeneratedWeek(user.ID.String())
			if err != nil {
				httpErr := app.HTTPError{
					Message: err.Error(),
					Code:    http.StatusInternalServerError,
				}
				return c.JSON(http.StatusInternalServerError, httpErr)
			}
			isLastGenerated = true
			for _, d := range week.Days {
				fmt.Printf("%s\n", d.Dinner.Name)
			}
		} else {
			httpErr := app.HTTPError{
				Message: "failed to fetch week: " + err.Error(),
				Code:    http.StatusInternalServerError,
			}
			return c.JSON(http.StatusInternalServerError, httpErr)
		}
	}

	allRecipes, err := uc.recipeService.Recipes()
	if err != nil {
		httpErr := app.HTTPError{
			Message: "failed to fetch recipes: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}

	newSuggestion := app.PickRecipeForDay(day, week.Days, allRecipes)

	if isLastGenerated {
		for _, d := range week.Days {
			if d.ID == day.ID {
				d.Dinner = newSuggestion
			}
		}
		_, err = uc.weekService.UpdateWeek(week, user.ID.String())
		if err != nil {
			httpErr := app.HTTPError{
				Message: "failed to update week: " + err.Error(),
				Code:    http.StatusInternalServerError,
			}
			return c.JSON(httpErr.Code, httpErr)
		}
	}

	return c.JSON(http.StatusOK, newSuggestion)

}

func (uc *UserController) UpdateWeek(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "failed to fetch user: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	week := &app.Week{}
	if err = c.Bind(week); err != nil {
		httpErr := app.HTTPError{
			Message: "failed to bind week: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	week, err = uc.weekService.UpdateWeek(week, user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: "failed to update week: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}

	return c.JSON(http.StatusOK, week)
}
func (uc *UserController) UpdateWeeks(c echo.Context) error {
	user, err := getUser(uc, c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "failed to fetch user: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	weeks := []*app.Week{}
	if err = c.Bind(&weeks); err != nil {
		httpErr := app.HTTPError{
			Message: "failed to bind weeks: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	weeks, err = uc.weekService.UpdateWeeks(weeks, user.ID.String())
	if err != nil {
		httpErr := app.HTTPError{
			Message: "failed to update weeks: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}

	return c.JSON(http.StatusOK, weeks)
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
