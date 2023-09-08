package app

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestGenerateDays(t *testing.T) {
	days := GenerateDays(2023, 1)
	if len(days) != 7 {
		t.Errorf("Expected 7 days, got %d", len(days))
	}
	expectedDates := []string{
		"2023-01-02",
		"2023-01-03",
		"2023-01-04",
		"2023-01-05",
		"2023-01-06",
		"2023-01-07",
		"2023-01-08",
	}
	for i, day := range days {
		if day.Date.Format("2006-01-02") != expectedDates[i] {
			t.Errorf("Expected date %s, got %s", expectedDates[i], day.Date)
		}
	}
}

func TestNoDuplicateRecipesInNearbyDays(t *testing.T) {
	rand.Seed(time.Now().Unix())
	days := createSomeDays(1000)
	recipes := createSomeRecipes(1000)

	// arrange
	for _, day := range days {
		recipe := PickRecipeForDay(day, days, recipes)
		day.Dinner = recipe
	}

	assertDinnerIsDifferent := func(day1, day2 *Day) {
		if day1.Dinner == day2.Dinner {
			t.Errorf("Expected dinner of nearby day to be different from dinner of today")
		}
	}

	// assert
	for i, day := range days {
		if i > 1 {
			assertDinnerIsDifferent(day, days[i-1])
		}
		if i < len(days)-1 {
			assertDinnerIsDifferent(day, days[i+1])
		}

	}
}

func createSomeDays(count int) []*Day {
	START_DATE := time.Now()

	days := make([]*Day, count)
	for i := range days {
		days[i] = &Day{
			Date: START_DATE.AddDate(0, 0, i),
		}
	}

	return days
}

func TestGenerateWeeks(t *testing.T) {
	seedTime, err := time.Parse("2006-01-02", "2023-12-14")
	if err != nil {
		t.Error(err)
	}
	expectedDays := []string{
		"50   2023-12-11   Monday",
		"50   2023-12-12   Tuesday",
		"50   2023-12-13   Wednesday",
		"50   2023-12-14   Thursday",
		"50   2023-12-15   Friday",
		"50   2023-12-16   Saturday",
		"50   2023-12-17   Sunday",
		"51   2023-12-18   Monday",
		"51   2023-12-19   Tuesday",
		"51   2023-12-20   Wednesday",
		"51   2023-12-21   Thursday",
		"51   2023-12-22   Friday",
		"51   2023-12-23   Saturday",
		"51   2023-12-24   Sunday",
		"52   2023-12-25   Monday",
		"52   2023-12-26   Tuesday",
		"52   2023-12-27   Wednesday",
		"52   2023-12-28   Thursday",
		"52   2023-12-29   Friday",
		"52   2023-12-30   Saturday",
		"52   2023-12-31   Sunday",
	}
	weeks := GenerateWeeks(seedTime)
	for i, week := range weeks {
		for j, day := range week.Days {
			index := i*7 + j
			_, weekNumber := day.Date.ISOWeek()
			res := fmt.Sprintf("%d   %s   %s", weekNumber, day.Date.Format("2006-01-02"), day.Date.Weekday())
			if res != expectedDays[index] {
				t.Errorf("Expected %s, got %s", expectedDays[index], res)
			}
		}
	}
}

func TestRecipeCost(t *testing.T) {
	items := []*Item{
		{Name: "Item 1", Price: 10, Amount: 10, Unit: "ml"},
		{Name: "Item 2", Price: 20, Amount: 4, Unit: "msk"},
		{Name: "Item 3", Price: 30, Amount: 1, Unit: "l"},
	}

	recipe := Recipe{
		NewRecipe: NewRecipe{
			Portions: 4,
			Items:    items,
		},
	}
	cost := recipe.Cost()
	if cost != 60 {
		t.Errorf("Expected cost 60, got %f", cost)
	}
}

func TestAlterRecipePortions(t *testing.T) {
	items := []*Item{
		{Name: "Item 1", Price: 10, Amount: 10, Unit: "ml"},
		{Name: "Item 2", Price: 20, Amount: 4, Unit: "msk"},
		{Name: "Item 3", Price: 30, Amount: 1, Unit: "l"},
	}

	recipe := &Recipe{
		NewRecipe: NewRecipe{
			Portions: 4,
			Items:    items,
		},
	}
	recipe = recipe.AlterPortions(2)
	if recipe.Portions != 2 {
		t.Errorf("Expected portions 2, got %d", recipe.Portions)
	}
	if recipe.Items[0].Amount != 5 {
		t.Errorf("Expected amount 5, got %f", recipe.Items[0].Amount)
	}
	if recipe.Items[1].Amount != 2 {
		t.Errorf("Expected amount 2, got %f", recipe.Items[1].Amount)
	}
	if recipe.Items[2].Amount != 0.5 {
		t.Errorf("Expected amount 0.5, got %f", recipe.Items[2].Amount)
	}
}

func TestAlterRecipePortionsEffectOnCost(t *testing.T) {
	items := []*Item{
		{Name: "Item 1", Price: 10, Amount: 10, Unit: "ml"},
		{Name: "Item 2", Price: 20, Amount: 4, Unit: "msk"},
		{Name: "Item 3", Price: 30, Amount: 1, Unit: "l"},
	}

	recipe := &Recipe{
		NewRecipe: NewRecipe{
			Name:     "Recipe 1",
			Portions: 4,
			Items:    items,
		},
	}
	cost := recipe.Cost()
	if cost != 60 {
		t.Errorf("Expected cost 60, got %f", cost)
	}
	recipe = recipe.AlterPortions(2)
	cost = recipe.Cost()
	if cost != 30 {
		t.Errorf("Expected cost 30, got %f", cost)
	}
}

func TestRandomSuggestions(t *testing.T) {
	recipes := createSomeRecipes(7)
	suggestionCount := 7 * 3
	suggestions := SuggestnRandomRecipes(recipes, suggestionCount)
	recCount := map[string]int{}
	for _, suggestion := range suggestions {
		recCount[suggestion.Name]++
	}
	for _, recipe := range recipes {
		rc := recCount[recipe.Name]
		expectedCount := int(math.Round(float64(suggestionCount) * recipe.ProbabilityWeight))
		// fmt.Printf("%s: %d < %d (%.1f * %.1f) = %t\n", recipe.Name, rc, expectedCount, float64(suggestionCount), recipe.ProbabilityWeight, (rc < expectedCount))
		if rc > expectedCount {
			t.Errorf("%s was suggested %d times, expected to be %f * %d = %d", recipe.Name, rc, recipe.ProbabilityWeight, suggestionCount, expectedCount)
		}
	}

	// repeatTest := 1000
	// testResults := map[string]int{}
	// for i := 0; i < repeatTest; i++ {
	// 	for i := 0; i < len(probs); i++ {
	// 		recipies = append(recipies, &Recipe{
	// 			NewRecipe: NewRecipe{
	// 				Name:              fmt.Sprintf("Recipe %d", i),
	// 				ProbabilityWeight: probs[i],
	// 			},
	// 		})
	// 	}
	// 	suggestions := suggestnRandomRecipes(recipies, suggestionCount)
	// 	recCount := map[string]int{}
	// 	for _, suggestion := range suggestions {
	// 		recCount[suggestion.Name]++
	// 	}
	// 	for i, recipe := range recipies {
	// 		rc := recCount[recipe.Name]
	// 		expectedCount := int(math.Round(float64(suggestionCount) * recipe.ProbabilityWeight))
	// 		if rc > expectedCount {
	// 			testResults[fmt.Sprintf("Recipe %d", i)]++
	// 		}
	// 	}
	// }
	// for i, recipe := range recipies {
	// 	expectedCount := int(math.Round(float64(repeatTest) * probs[i]))
	// 	count := testResults[recipe.Name]
	// 	if count > expectedCount {
	// 		t.Errorf("%s was suggested %d times, expected to be %f * %d = %d", recipe.Name, count, probs[i], repeatTest, expectedCount)
	// 	}
	// }
}

func createSomeRecipes(count int) []*Recipe {
	weights := make([]float64, count)
	for i := range weights {
		weights[i] = rand.Float64()
	}

	recipes := make([]*Recipe, count)
	for i := range recipes {
		recipes[i] = &Recipe{
			NewRecipe: NewRecipe{
				Name:              fmt.Sprintf("Recipe %d", i),
				ProbabilityWeight: weights[i],
			},
		}
	}
	return recipes
}
