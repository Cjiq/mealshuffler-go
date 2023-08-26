package app

import (
	"math/rand"
	"sort"
	"time"

	"github.com/snabb/isoweek"
)

type Entity struct {
	ID int64 `json:"id,omitempty"`
}

type NewUser struct {
	Name string `json:"name,omitempty"`
}

type User struct {
	Weeks []*Week `json:"weeks,omitempty"`
	NewUser
	Entity
}

type NewRecipe struct {
	Name              string  `json:"name,omitempty"`
	Items             []*Item `json:"items,omitempty"`
	ProbabilityWeight float64 `json:"probability_weight,omitempty"`
	Portions          int     `json:"portions,omitempty"`
}

type Recipe struct {
	NewRecipe
	Entity
}

type Item struct {
	Name   string  `json:"name,omitempty"`
	Price  int     `json:"price,omitempty"`
	Amount float64 `json:"amount,omitempty"`
	Unit   string  `json:"unit,omitempty"`
	Entity
}

type Week struct {
	Days [7]*Day `json:"days,omitempty"`
	Entity
}

type Day struct {
	Date   time.Time `json:"date,omitempty"`
	Dinner *Recipe   `json:"dinner,omitempty"`
	Lunch  *Recipe   `json:"lunch,omitempty"`
	Entity
}

type UserService interface {
	User(id int) (*User, error)
	Users() ([]*User, error)
	CreateUser(u *NewUser) (*User, error)
	DeleteUser(id int) error
}
type RecipeService interface {
	Recipe(id int) (*Recipe, error)
	Recipes() ([]*Recipe, error)
	CreateRecipe(u *Recipe) (*Recipe, error)
	DeleteRecipe(id int) error
	UserRecipes(userID int) ([]*Recipe, error)
}
type ItemService interface {
	Item(id int) (*Item, error)
	Items() ([]*Item, error)
	CreateItem(u *Item) (*Item, error)
	DeleteItem(id int) error
}

func (r *Recipe) AlterPortions(portions int) *Recipe {
	alteredItems := make([]*Item, len(r.Items))
	for i, item := range r.Items {
		alteredFraction := float64(portions) / float64(r.Portions)
		alteredItems[i] = &Item{
			Name:   item.Name,
			Price:  int(float64(item.Price) * alteredFraction),
			Amount: item.Amount * alteredFraction,
			Unit:   item.Unit,
		}
	}
	r.Items = alteredItems
	r.Portions = portions
	return r
}

func (r *Recipe) Cost() float64 {
	var cost float64
	for _, item := range r.Items {
		cost += float64(item.Price)
	}
	return cost
}

// GenerateDays generates a slice of days for a given year and iso week number.
func GenerateDays(year, weekNumber int) [7]*Day {

	startDate := isoweek.StartTime(year, weekNumber, time.UTC)

	// Generate dates for each day of the week.
	var days [7]*Day
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		days[i] = &Day{Date: date}
	}

	return days
}

// GenerateWeeks generates a slice of weeks for the rest of the year
// based on the current date.
func GenerateWeeks(startTime time.Time) []*Week {
	weeks := make([]*Week, 0)

	now := startTime
	year, week := now.ISOWeek()

	for i := week; i <= 52; i++ {
		weeks = append(weeks, &Week{
			Days: GenerateDays(year, i),
		})
	}
	return weeks
}

func suggestnRandomRecipes(recipes []*Recipe, n int) []*Recipe {
	var selectedRecipes []*Recipe
	lastPickedRecipes := make([]*Recipe, 0, 5)

	for i := 0; i < n; i++ {
		// Sort recipes by probability in descending order
		sort.Slice(recipes, func(i, j int) bool {
			return recipes[i].ProbabilityWeight > recipes[j].ProbabilityWeight
		})

		// Filter recipes that were picked in the last 5 selections
		availableRecipes := make([]*Recipe, 0, len(recipes))
		for _, recipe := range recipes {
			if !containsRecipe(lastPickedRecipes, recipe) {
				availableRecipes = append(availableRecipes, recipe)
			}
		}

		// Choose a recipe based on probabilities
		chosenRecipe := chooseRecipe(availableRecipes)

		// Add the chosen recipe to the selected recipes and update lastPickedRecipes
		selectedRecipes = append(selectedRecipes, chosenRecipe)
		lastPickedRecipes = append(lastPickedRecipes, chosenRecipe)
		if len(lastPickedRecipes) > 5 {
			lastPickedRecipes = lastPickedRecipes[1:]
		}
	}

	return selectedRecipes
}

func containsRecipe(recipes []*Recipe, target *Recipe) bool {
	for _, recipe := range recipes {
		if recipe.Name == target.Name {
			return true
		}
	}
	return false
}

func chooseRecipe(recipes []*Recipe) *Recipe {
	totalWeight := 0.0
	for _, recipe := range recipes {
		totalWeight += recipe.ProbabilityWeight
	}

	r := rand.Float64() * totalWeight
	cumulativeWeight := 0.0

	for _, recipe := range recipes {
		cumulativeWeight += recipe.ProbabilityWeight
		if r <= cumulativeWeight {
			return recipe
		}
	}

	return recipes[len(recipes)-1]
}
