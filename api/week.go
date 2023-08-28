package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type WeekController struct {
	weekService app.WeekService
}

func NewWeekController(weekService app.WeekService) *WeekController {
	return &WeekController{weekService: weekService}
}

func (wc *WeekController) GetWeeks(c echo.Context) error {
	id := c.Param("id")
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	weeks, err := wc.weekService.Weeks(id, year)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	return c.JSON(http.StatusOK, weeks)
}

func (wc *WeekController) CreateWeek(c echo.Context) error {
	userID := c.Param("id")
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

	week, err := wc.weekService.CreateWeek(&newWeek, userID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, week)
}

func (wc *WeekController) CreateWeeks(c echo.Context) error {
	userID := c.Param("id")
	newWeeks := []*app.NewWeek{}
	if err := c.Bind(&newWeeks); err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}

	weeks, err := wc.weekService.CreateWeeks(newWeeks, userID)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	return c.JSON(http.StatusOK, weeks)
}
