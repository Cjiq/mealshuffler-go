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
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, recipes)
}

func (rc *RecipeController) CreateRecipe(c echo.Context) error {
	var newRecipe app.NewRecipe
	var json map[string]interface{}
	if err := c.Bind(&json); err != nil {
		return c.String(http.StatusBadRequest, "Error: "+err.Error())
	}
	if json["name"] == nil {
		httpErr := app.HTTPError{
			Message: "name is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	} else {
		newRecipe.Name = json["name"].(string)
	}
	if json["probability_weight"] == nil {
		httpErr := app.HTTPError{
			Message: "probability_weight is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	} else {
		newRecipe.ProbabilityWeight = json["probability_weight"].(float64)
	}
	if json["portions"] == nil {
		httpErr := app.HTTPError{
			Message: "portions is required",
			Code:    http.StatusUnprocessableEntity,
		}
		return c.JSON(http.StatusUnprocessableEntity, httpErr)
	} else {
		newRecipe.Portions = int(json["portions"].(float64))
	}

	recipe, err := rc.recipeService.CreateRecipe(&newRecipe)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error: "+err.Error())
	}
	return c.JSON(http.StatusOK, recipe)
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
