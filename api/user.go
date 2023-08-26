package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type UserController struct {
	userService app.UserService
}

func NewUserController(userService app.UserService) *UserController {
	return &UserController{userService: userService}
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
	id := c.Param("id")
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusBadRequest, "ID must be an integer")
	}
	user, err := uc.userService.User(idNum)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error")
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
