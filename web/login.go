package web

import (
	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
	"nrdev.se/mealshuffler/web/views"
)

func LoginIndex(c echo.Context, userService app.UserService) error {
	component := views.Login()
	return RenderWithLayout(c, component)
}
