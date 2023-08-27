package sqlite

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"nrdev.se/mealshuffler/app"
)

type DayService struct {
	db *sql.DB
}

func NewDayService(db *sql.DB) *DayService {
	us := &DayService{db: db}
	us.CreateDayTable()
	return &DayService{db: db}
}

func (ds *DayService) CreateDayTable() error {
	query := `CREATE TABLE IF NOT EXISTS day (
		id TEXT PRIMARY KEY,
		date TEXT NOT NULL,
		week_id TEXT NOT NULL,
		dinner_id TEXT NOT NULL
	);`
	if _, err := ds.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (ds *DayService) Days() ([]*app.Day, error) {
	rows, err := ds.db.Query(`SELECT d.id, d.date,
	r.id, r.name, r.probability_weight, r.portions
	FROM day as d
	INNER JOIN recipe as r ON r.id = d.dinner_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	days := make([]*app.Day, 0)
	for rows.Next() {
		var dateStr string
		var r app.Recipe
		var d app.Day
		if err := rows.Scan(&d.ID, &dateStr, &r.ID, &r.Name, &r.ProbabilityWeight, &r.Portions); err != nil {
			return nil, err
		}
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		d.Date = date
		d.Dinner = &r
		days = append(days, &d)
	}

	return days, nil
}

func (ds *DayService) CreateDay(newDay *app.NewDay) (*app.Day, error) {
	id := uuid.New()
	tx, err := ds.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO day(id, date, dinner_id) VALUES(?,?,?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id.String(), newDay.Date, newDay.Dinner.ID)
	if err != nil {
		return nil, err
	}

	day := &app.Day{
		Entity: app.Entity{
			ID: id,
		},
		NewDay: app.NewDay{
			Date:   newDay.Date,
			Dinner: newDay.Dinner,
		},
	}
	tx.Commit()

	return day, nil
}
