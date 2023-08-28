package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type WeekController struct {
	weekService app.WeekService
}

func NewWeekController(weekService app.WeekService) *WeekController {
	return &WeekController{weekService: weekService}
}

func (rc *WeekController) GetWeeks(c echo.Context) error {
	weeks, err := rc.weekService.Weeks()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, weeks)
}

func (rc *WeekController) CreateWeek(c echo.Context) error {
	newWeek := app.NewWeek{}
	if err := c.Bind(&newWeek); err != nil {
		return c.String(http.StatusBadRequest, "Error: "+err.Error())
	}
	if newWeek.Number == 0 {
		httpErr := app.HTTPError{
			Message: "number need to be set between 1 and 52",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}

	if len(newWeek.Days) == 0 {
		httpErr := app.HTTPError{
			Message: "days need to be set",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}

	week, err := rc.weekService.CreateWeek(&newWeek)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, week)
}
