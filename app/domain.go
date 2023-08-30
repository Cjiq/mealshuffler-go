package app

import (
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/snabb/isoweek"
)

type Entity struct {
	ID uuid.UUID `json:"id,omitempty"`
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
	URL               string  `json:"url,omitempty"`
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

type NewWeek struct {
	Days   []*Day `json:"days,omitempty"`
	Number int    `json:"number,omitempty"`
	Year   int    `json:"year,omitempty"`
}

type Week struct {
	NewWeek
	Entity
}

type Day struct {
	Date   time.Time `json:"date,omitempty"`
	Dinner *Recipe   `json:"dinner,omitempty"`
	Entity
}

type HTTPError struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type UserService interface {
	User(id string) (*User, error)
	Users() ([]*User, error)
	CreateUser(u *NewUser) (*User, error)
	DeleteUser(id string) error
	UserWeeks(userID string, startWeek, skipWeek int) ([]*Week, error)
	SaveUserWeeks(userID string, weeks []*Week) error
}
type RecipeService interface {
	// Recipe(id int) (*Recipe, error)
	Recipes() ([]*Recipe, error)
	CreateRecipe(rs *NewRecipe) (*Recipe, error)
	UpdateRecipe(rs *Recipe) (*Recipe, error)
	// DeleteRecipe(id int) error
	// UserRecipes(userID int) ([]*Recipe, error)
	DeleteAllRecipes() error
}

//	type ItemService interface {
//		Item(id int) (*Item, error)
//		Items() ([]*Item, error)
//		CreateItem(u *Item) (*Item, error)
//		DeleteItem(id int) error
//	}
type WeekService interface {
	Week(id string, userID string) (*Week, error)
	Weeks(userID string, year int) ([]*Week, error)
	CreateWeek(w *NewWeek, userID string) (*Week, error)
	CreateWeeks(w []*NewWeek, userID string) ([]*Week, error)
	DeleteWeek(id string, userID string) error
	UpdateWeek(w *Week, userID string) (*Week, error)
	LastGeneratedWeek(userID string) (*Week, error)
	DeleteWeeks(userID string, year int) error
	NextWeekNumber(userID string) (int, error)
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
func GenerateDays(year, weekNumber int) []*Day {

	startDate := isoweek.StartTime(year, weekNumber, time.UTC)

	// Generate dates for each day of the week.
	var days []*Day
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		d := &Day{
			Date: date,
		}
		days = append(days, d)
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
			NewWeek: NewWeek{
				Days: GenerateDays(year, i),
			},
		})
	}
	return weeks
}

func SuggestnRandomRecipes(recipes []*Recipe, n int) []*Recipe {
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
