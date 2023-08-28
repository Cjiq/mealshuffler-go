package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"nrdev.se/mealshuffler/app"
)

type WeekService struct {
	db *sql.DB
}

func NewWeekService(db *sql.DB) *WeekService {
	ws := &WeekService{db: db}
	err := ws.CreateWeekTable()
	if err != nil {
		panic(err)
	}
	return ws
}

func (ws *WeekService) CreateWeekTable() error {
	query := (`
		CREATE TABLE IF NOT EXISTS weeks (
			id TEXT PRIMARY KEY,
			days TEXT,
			number INTEGER NOT NULL,
			user_id TEXT NOT NULL,
		);
	`)
	if _, err := ws.db.Exec(query); err != nil {
		return err
	}
	return nil
}

func (ws *WeekService) Week(id string) (*app.Week, error) {
	query := (`
		SELECT id, days, number
		FROM weeks
		WHERE id = ?
	`)
	row := ws.db.QueryRow(query, id)
	var week app.Week
	var daysJSON string
	err := row.Scan(&week.ID, &daysJSON, &week.Number)
	if err != nil {
		return nil, err
	}
	var days []*app.Day
	err = json.Unmarshal([]byte(daysJSON), &days)
	if err != nil {
		return nil, err
	}
	return &week, nil
}

func (ws *WeekService) Weeks(userID string, year int) ([]*app.Week, error) {
	query := (`
		SELECT id, days, number, year
		FROM weeks
		WHERE user_id = ?
	`)
	rows, err := ws.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var weeks []*Week
	for rows.Next() {
		var week Week
		var daysJSON string
		err := rows.Scan(&week.ID, &daysJSON, &week.Number)
		if err != nil {
			return nil, err
		}
		var days []*app.Day
		err = json.Unmarshal([]byte(daysJSON), &days)
		if err != nil {
			return nil, err
		}
		week.Days = days
		weeks = append(weeks, &week)
	}
	return weeks, nil
}

func (ws *WeekService) CreateWeek(newWeek *app.NewWeek, userID string) (*app.Week, error) {
	id := uuid.New()
	query := (`
		INSERT INTO weeks (id, days, number, user_id)
		VALUES (?, ?, ?, ?)
	`)
	daysJSON, err := json.Marshal(newWeek.Days)
	if err != nil {
		return nil, err
	}
	tx, err := ws.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec(id, daysJSON, newWeek.Number, userID.String())
	if err != nil {
		return nil, err
	}
	return &app.Week{
		Entity: app.Entity{
			ID: id,
		},
		NewWeek: *newWeek,
	}, nil
}

func (ws *WeekService) CreateWeeks(newWeeks []*app.NewWeek, userID uuid.UUID) ([]*app.Week, error) {
	weeks := []*app.Week{}
	tx, err := ws.db.Begin()
	if err != nil {
		return nil, err
	}
	query := (`
		INSERT INTO weeks (id, days, number, user_id)
		VALUES (?, ?, ?, ?)
	`)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	for _, newWeek := range newWeeks {
		var week *app.Week
		var daysJSON []byte
		daysJSON, err = json.Marshal(newWeek.Days)
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(uuid.New(), string(daysJSON), newWeek.Number, userID.String())
		if err != nil {
			return nil, err
		}
		weeks = append(weeks, week)
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return weeks, nil
}
