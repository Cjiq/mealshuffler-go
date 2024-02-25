package web

import (
	"log"

	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
	"nrdev.se/mealshuffler/web/views"
)

func Index(c echo.Context, userService app.UserService) error {
	users, err := userService.Users()
	if err != nil {
		log.Fatal("U SUCK DUUH, ", err)
	}
	component := views.Home("Hello", users)
	return RenderWithLayout(c, component)
}
