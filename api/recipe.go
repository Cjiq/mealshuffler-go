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
