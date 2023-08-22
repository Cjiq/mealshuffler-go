package app

import (
	"time"

	"github.com/snabb/isoweek"
)

type NewUser struct {
	Name string `json:"name,omitempty"`
}

type User struct {
	ID int64 `json:"id,omitempty"`
	NewUser
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
}
type ItemService interface {
	Item(id int) (*Item, error)
	Items() ([]*Item, error)
	CreateItem(u *Item) (*Item, error)
	DeleteItem(id int) error
}

type Recipe struct {
	Name  string  `json:"name,omitempty"`
	Items []*Item `json:"items,omitempty"`
}

type Item struct {
	Name  string `json:"name,omitempty"`
	Price string `json:"price,omitempty"`
}

type Week struct {
	Days [7]*Day `json:"days,omitempty"`
}

type Day struct {
	Date   time.Time `json:"date,omitempty"`
	Dinner *Recipe   `json:"dinner,omitempty"`
	Lunch  *Recipe   `json:"lunch,omitempty"`
}

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
