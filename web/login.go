package web

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"nrdev.se/mealshuffler/app"
	"nrdev.se/mealshuffler/web/views"
)

type LoginHandler struct {
	UserService app.UserService
}

func (handler *LoginHandler) Index(c echo.Context) error {
	component := views.Login()
	return RenderWithLayout(c, component)
}

func (handler *LoginHandler) Do(c echo.Context) error {
	type user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var u user
	if err := c.Bind(&u); err != nil {
		httpErr := app.HTTPError{
			Message:  err.Error(),
			Friendly: "Failed to parse request",
			Code:     http.StatusBadRequest,
		}
		return Render(c, views.Errors(httpErr))
	}

	userHash, err := handler.UserService.GetUserHash(u.Username)
	if err != nil {
		httpErr := app.HTTPError{
			Message:  err.Error(),
			Friendly: "invalid username or password",
			Code:     http.StatusInternalServerError,
		}
		return Render(c, views.Errors(httpErr))
	}
	err = bcrypt.CompareHashAndPassword(userHash, []byte(u.Password))
	if err != nil {
		httpErr := app.HTTPError{
			Message:  "invalid username or password",
			Friendly: "invalid username or password",
			Code:     http.StatusBadRequest,
		}
		return Render(c, views.Errors(httpErr))
	}
	sess, err := session.Get("auth-session", c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return Render(c, views.Errors(httpErr))
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values["authenticated"] = true
	sess.Save(c.Request(), c.Response())
	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func (handler *LoginHandler) Undo(c echo.Context) error {
	sess, err := session.Get("auth-session", c)
	if err != nil {
		httpErr := app.HTTPError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return Render(c, views.Errors(httpErr))
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7 * -1,
		HttpOnly: true,
	}
	sess.Values["authenticated"] = true
	sess.Save(c.Request(), c.Response())
	return c.Redirect(303, "/login")
}
