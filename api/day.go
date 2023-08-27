package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type DayController struct {
	dayService app.DayService
}

func NewDayController(dayService app.DayService) *DayController {
	return &DayController{dayService: dayService}
}

func (dc *DayController) GetDays(c echo.Context) error {
	days, err := dc.dayService.Days()
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	return c.JSON(http.StatusOK, days)
}
