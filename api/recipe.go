package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"nrdev.se/mealshuffler/app"
)

type RecipeController struct {
	recipeService app.RecipeService
}

func NewRecipeController(recipeService app.RecipeService) *RecipeController {
	return &RecipeController{recipeService: recipeService}
}

func (rc *RecipeController) GetRecipes(c echo.Context) error {
	recipes, err := rc.recipeService.Recipes()
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	return c.JSON(http.StatusOK, recipes)
}

func (rc *RecipeController) CreateRecipe(c echo.Context) error {
	var newRecipe app.NewRecipe
	if err := c.Bind(&newRecipe); err != nil {
		return c.String(http.StatusBadRequest, "Error: "+err.Error())
	}
	if newRecipe.Name == "" {
		httpErr := app.HTTPError{
			Message: "name is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}
	if newRecipe.ProbabilityWeight == 0 {
		httpErr := app.HTTPError{
			Message: "probability_weight is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}
	if newRecipe.Portions == 0 {
		httpErr := app.HTTPError{
			Message: "portions is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}
	userID := c.Param("id")
	if userID == "" {
		httpErr := app.HTTPError{
			Message: "userID is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(httpErr.Code, httpErr)
	}

	recipe, err := rc.recipeService.CreateRecipe(&newRecipe, userID)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(httpErr.Code, httpErr)
	}
	c.Response().Header().Set("Location", "/recipes/"+recipe.ID.String())
	return c.JSON(http.StatusCreated, recipe)
}

func (rc *RecipeController) DeleteRecipes(c echo.Context) error {
	err := rc.recipeService.DeleteAllRecipes()
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	return c.NoContent(http.StatusNoContent)
}

func (rc *RecipeController) UpdateRecipe(c echo.Context) error {
	var recipe app.Recipe
	if err := c.Bind(&recipe); err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusBadRequest,
		}
		return c.JSON(http.StatusBadRequest, httpErr)
	}
	if recipe.Name == "" {
		httpErr := app.HTTPError{
			Message: "Error: name is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}
	if recipe.ProbabilityWeight == 0 {
		httpErr := app.HTTPError{
			Message: "Error: probability_weight is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}
	if recipe.Portions == 0 {
		httpErr := app.HTTPError{
			Message: "Error: portions is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	}
	updatedRecipe, err := rc.recipeService.UpdateRecipe(&recipe)
	if err != nil {
		httpErr := app.HTTPError{
			Message: "Error: " + err.Error(),
			Code:    http.StatusInternalServerError,
		}
		return c.JSON(http.StatusInternalServerError, httpErr)
	}
	return c.JSON(http.StatusOK, updatedRecipe)
}
