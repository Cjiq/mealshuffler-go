package web

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/web/views"
)

func Render(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
func RenderWithLayout(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return views.Layout(cmp).Render(c.Request().Context(), c.Response().Writer)
}
