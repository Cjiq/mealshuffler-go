package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"nrdev.se/mealshuffler/api"
	"nrdev.se/mealshuffler/app"
	"nrdev.se/mealshuffler/web"
)

func (srv *server) setupRoutes(e *echo.Echo) {

	rc := api.NewRecipeController(srv.recipeService)
	uc := api.NewUserController(srv.userService, srv.recipeService, srv.weekService)
	wc := api.NewWeekController(srv.weekService)

	// API ROUTES
	api := e.Group("/api")
	api.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(key string, _ echo.Context) (bool, error) {
			_, err := srv.userService.ValidateUserToken(key)
			if err != nil {
				log.Println("error validating token: ", err)
				return false, err
			}
			return true, nil
		},
		ErrorHandler: func(err error, _ echo.Context) error {
			httpErr := app.HTTPError{
				Message: fmt.Sprintf("access denied: %s", err.Error()),
				Code:    http.StatusUnauthorized,
			}
			return httpErr
		},
	}))

	api.GET("/ping", ping)

	api.GET("/users/:id/generate", uc.GenerateWeek)
	api.POST("/users/:id/weeks", uc.SaveWeek)
	api.DELETE("/users/:id/weeks/:weekID", uc.DeleteWeek)
	api.GET("/users/:id/weeks/next", uc.NextWeekNumber)
	api.POST("/users/:id/weeks/:weekID/suggest", uc.GenerateRecipeAlternative)
	api.PUT("/users/:id/weeks/:weekID", uc.UpdateWeek)
	api.PUT("/users/:id/weeks", uc.UpdateWeeks)
	api.PUT("/users/:id/weeks/shuffle", uc.ShuffleWeekRecipes)
	api.POST("/users/:id/recipes", rc.CreateRecipe)

	api.GET("/users/:id/recipes", rc.GetUserRecipes)

	api.GET("/recipes", rc.GetRecipes)
	api.DELETE("/recipes", rc.DeleteRecipes)
	api.PUT("/recipes", rc.UpdateRecipe)

	api.GET("/users/:id/weeks/:year", wc.GetWeeks)
	api.GET("/users/:id/weeks/last", wc.GetLastGeneratedWeek)
	api.DELETE("/users/:id/weeks/:year/all", wc.DeleteWeeks)

	admin := e.Group("/api/admin")
	admin.Use(srv.AdminMiddleware)
	admin.GET("/ping", ping)
	admin.POST("/users", uc.CreateUser)
	admin.GET("/users", uc.GetUsers)
	admin.GET("/users/:id", uc.GetUser)
	admin.DELETE("/users/:id", uc.DeleteUser)

	login := &web.LoginHandler{
		UserService: srv.userService,
	}

	// WEB ROUTES
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("auth-session"))))
	e.POST("/login", srv.login)
	e.GET("/", func(c echo.Context) error {
		sess, _ := session.Get("auth-session", c)
		if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
			return c.Redirect(303, "/login")
		} else {
			return web.Index(c, srv.userService)
		}

	})
	e.GET("/login", login.Index)
	e.POST("/login/do", login.Do)
	e.GET("/login/undo", login.Undo)
	e.POST("/login/undo", login.Undo)

}

func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}
