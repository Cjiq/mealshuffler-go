package app_test

import (
	"fmt"
	"testing"
	"time"

	"nrdev.se/mealshuffler/app"
)

func TestGenerateDays(t *testing.T) {
	days := app.GenerateDays(2023, 1)
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
	weeks := app.GenerateWeeks(seedTime)
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
