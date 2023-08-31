package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

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
		CREATE TABLE IF NOT EXISTS week (
			id TEXT PRIMARY KEY,
			days TEXT,
			number INTEGER NOT NULL,
			year INTEGER NOT NULL,
			user_id TEXT NOT NULL
		);
	`)
	if _, err := ws.db.Exec(query); err != nil {
		return err
	}
	return nil
}

func (ws *WeekService) Week(id string, userID string) (*app.Week, error) {
	query := (`
		SELECT id, days, number
		FROM week
		WHERE id = ? AND user_id = ?
	`)
	row := ws.db.QueryRow(query, id, userID)
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
	week.Days = days
	return &week, nil
}

func (ws *WeekService) Weeks(userID string, year int) ([]*app.Week, error) {
	query := (`
		SELECT DISTINCT id, days, number, year
		FROM week
		WHERE user_id = ? AND year = ? and number != -1
	`)
	rows, err := ws.db.Query(query, userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	weeks := []*app.Week{}
	for rows.Next() {
		week := &app.Week{}
		var daysJSON string
		err := rows.Scan(&week.ID, &daysJSON, &week.Number, &week.Year)
		if err != nil {
			return nil, err
		}
		var days []*app.Day
		err = json.Unmarshal([]byte(daysJSON), &days)
		if err != nil {
			return nil, err
		}
		week.Days = days
		weeks = append(weeks, week)
	}
	return weeks, nil
}

func (ws *WeekService) CreateWeek(newWeek *app.NewWeek, userID string) (*app.Week, error) {
	id := uuid.New()
	query := (`
		INSERT INTO week (id, days, number, year, user_id)
		VALUES (?, ?, ?, ?, ?)
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
	_, err = stmt.Exec(id, daysJSON, newWeek.Number, newWeek.Year, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
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

func (ws *WeekService) UpdateWeek(week *app.Week, userID string) (*app.Week, error) {
	query := (`
		UPDATE week
		SET days = ?, number = ?, year = ?
		WHERE id = ? AND user_id = ?
	`)
	daysJSON, err := json.Marshal(week.Days)
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
	res, err := stmt.Exec(daysJSON, week.Number, week.Year, week.ID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("week not found")
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return week, nil
}

func (ws *WeekService) CreateWeeks(newWeeks []*app.NewWeek, userID string) ([]*app.Week, error) {
	weeks := []*app.Week{}
	tx, err := ws.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query := (`
		INSERT INTO week (id, days, number, year, user_id)
		VALUES (?, ?, ?, ?, ?)
	`)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	for _, newWeek := range newWeeks {
		var week *app.Week
		var daysJSON []byte
		daysJSON, err = json.Marshal(newWeek.Days)
		id := uuid.New()
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(id, string(daysJSON), newWeek.Number, newWeek.Year, userID)
		if err != nil {
			return nil, err
		}
		week = &app.Week{
			Entity: app.Entity{
				ID: id,
			},
			NewWeek: *newWeek,
		}
		weeks = append(weeks, week)
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return weeks, nil
}

func (ws *WeekService) LastGeneratedWeek(userID string) (*app.Week, error) {
	query := (`
		SELECT id, days, number, year
		FROM week
		WHERE number = -1 AND user_id = ?
	`)
	row := ws.db.QueryRow(query, userID)
	var week app.Week
	var daysJSON string
	err := row.Scan(&week.ID, &daysJSON, &week.Number, &week.Year)
	if err != nil {
		return nil, err
	}
	var days []*app.Day
	err = json.Unmarshal([]byte(daysJSON), &days)
	if err != nil {
		return nil, err
	}
	week.Days = days
	return &week, nil
}

func (ws *WeekService) DeleteWeeks(userID string, year int) error {
	query := (`
		DELETE FROM week
		WHERE user_id = ? AND year = ?
	`)
	tx, err := ws.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(userID, year)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (ws *WeekService) DeleteWeek(id string, userID string) error {
	query := (`
		DELETE FROM week
		WHERE id = ? AND user_id = ?
	`)
	tx, err := ws.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id, userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (ws *WeekService) NextWeekNumber(userID string) (int, error) {
	query := (`
		SELECT MAX(number)
		FROM week
		WHERE user_id = ?
	`)
	row := ws.db.QueryRow(query, userID)
	var number int
	err := row.Scan(&number)
	if err != nil {
		return 0, err
	}
	return number + 1, nil
}
