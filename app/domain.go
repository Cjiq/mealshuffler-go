package app

import (
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/snabb/isoweek"
)

type Entity struct {
	ID uuid.UUID `json:"id,omitempty"`
}

type NewUser struct {
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type User struct {
	Weeks []*Week `json:"weeks,omitempty"`
	token string
	NewUser
	Entity
}

type NewRecipe struct {
	Name               string  `json:"name,omitempty"`
	Items              []*Item `json:"items,omitempty"`
	ProbabilityWeight  float64 `json:"probability_weight,omitempty"`
	Portions           int     `json:"portions,omitempty"`
	URL                string  `json:"url,omitempty"`
	LeftOverCompliance bool    `json:"left_over_compliance,omitempty"`
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
	CreateUser(u *NewUser, hash []byte) (*User, error)
	DeleteUser(id string) error
	UserWeeks(userID string, startWeek, skipWeek int) ([]*Week, error)
	// SaveUserWeeks(userID string, weeks []*Week) error
	ValidateUserToken(token string) (string, error)
	SaveUserToken(userID string, token string) error
	GetUserHash(userID string) ([]byte, error)
	UserByUserName(username string) (*User, error)
	GetUserToken(userID string) (string, error)
}
type RecipeService interface {
	// Recipe(id int) (*Recipe, error)
	Recipes() ([]*Recipe, error)
	CreateRecipe(rs *NewRecipe, userID string) (*Recipe, error)
	UpdateRecipe(rs *Recipe) (*Recipe, error)
	UserRecipes(userID string) ([]*Recipe, error)
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
	UpdateWeeks(w []*Week, userID string) ([]*Week, error)
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

type DayWithRecipeSelectionMetadata struct {
	Day                                  *Day
	DistanceInDaysToDayToSelectRecipeFor int
}

type RecipeWithSelectionMetadata struct {
	Recipe          *Recipe
	SelectionWeight float64
}

// todo: don't always pick the first recipe
func PickRecipeForDay(dayToSelectRecipeFor *Day, context []*Day, allRecipes []*Recipe) *Recipe {
	if len(allRecipes) == 0 {
		panic("Cannot pick recipe from empty slice (allRecipes)")
	}
	daysWithMetadata := make([]*DayWithRecipeSelectionMetadata, len(context))
	for i, day := range context {
		daysWithMetadata[i] = newRecipeSelectionMetadataForDay(day, dayToSelectRecipeFor)
	}

	getSelectionWeightMultiplier := func(recipe *Recipe) float64 {
		const MAX_DIST_DAYS = 5
		for _, day := range daysWithMetadata {
			if day.Day.Dinner == nil {
				continue
			}
			if day.Day.Dinner.Name == recipe.Name && day.DistanceInDaysToDayToSelectRecipeFor <= MAX_DIST_DAYS {
				return 0
			}
		}
		return 1
	}

	recipesWithMetadata := make([]*RecipeWithSelectionMetadata, len(allRecipes))
	for i, recipe := range allRecipes {
		recipesWithMetadata[i] = &RecipeWithSelectionMetadata{
			Recipe:          recipe,
			SelectionWeight: getSelectionWeightMultiplier(recipe) * recipe.ProbabilityWeight,
		}
	}

	selectedRecipe := PickRandom(recipesWithMetadata, func(recipe *RecipeWithSelectionMetadata) float64 {
		return recipe.SelectionWeight
	})
	return selectedRecipe.Recipe
}

func newRecipeSelectionMetadataForDay(day *Day, dayToSelectRecipeFor *Day) *DayWithRecipeSelectionMetadata {
	value := &DayWithRecipeSelectionMetadata{
		Day:                                  day,
		DistanceInDaysToDayToSelectRecipeFor: GetAbsoluteTimeDifferenceInDays(day.Date, dayToSelectRecipeFor.Date),
	}
	return value
}

func GetAbsoluteTimeDifferenceInDays(a time.Time, b time.Time) int {
	difference := a.UTC().Sub(b.UTC())
	distanceInHours := math.Abs(difference.Hours())
	return int(distanceInHours / 24)
}

func containsRecipe(recipes []*Recipe, target *Recipe) bool {
	for _, recipe := range recipes {
		if recipe.Name == target.Name {
			return true
		}
	}
	return false
}
