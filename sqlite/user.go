package sqlite

import (
	"database/sql"

	"github.com/google/uuid"

	"nrdev.se/mealshuffler/app"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	us := &UserService{db: db}
	us.CreateUserTable()
	return us
}

func (u *UserService) CreateUserTable() error {
	query := `CREATE TABLE IF NOT EXISTS user (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL
	);`
	if _, err := u.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (u *UserService) Users() ([]*app.User, error) {
	rows, err := u.db.Query("SELECT id, name FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*app.User, 0)
	for rows.Next() {
		var u app.User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, nil
}

func (u *UserService) CreateUser(newUser *app.NewUser) (*app.User, error) {
	id := uuid.New()
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO user(name, id) VALUES(?,?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newUser.Name, id.String())
	if err != nil {
		return nil, err
	}

	user := &app.User{
		Entity: app.Entity{
			ID: id,
		},
		NewUser: app.NewUser{
			Name: newUser.Name,
		},
	}
	tx.Commit()

	return user, nil
}

func (u *UserService) User(id string) (*app.User, error) {
	var idStr string
	var user app.User
	if err := u.db.QueryRow("SELECT id, name FROM user WHERE id = ?", id).Scan(&idStr, &user.Name); err != nil {
		return nil, err
	}
	var err error
	user.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserService) DeleteUser(id int) error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("DELETE FROM user WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (us *UserService) UserWeeks(id string, startWeek int, weekCount int) ([]*app.Week, error) {
	endWeek := startWeek + weekCount - 1
	if endWeek > 52 {
		endWeek = 52
	}
	rows, err := us.db.Query(`SELECT 
		w.id, w.number,
		d.id, d.date, 
		r.id, r.name, r.portions, r.probability_weight
	FROM week as w 
	INNER JOIN day as d ON d.week_id = w.id
	INNER JOIN recipe as r ON r.id = d.dinner_id
	WHERE 
		user_id = ?
		AND w.number >= ?
		AND w.number <= ?`, id, startWeek, endWeek)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	weeks := make([]*app.Week, 0)
	var days []*app.Day
	lastWeek := -1
	w := &app.Week{}
	d := &app.Day{}
	r := &app.Recipe{}
	var wIDStr, dIDStr, rIDStr, weekDateStr string
	// wID, wNumber, dID, dDate, rID, rName, rPortions, rProbabilityWeight
	// -------------------------------------------------------------------
	// 0,   1,       1,   8,      4,   5,     6,         7
	// 0,   1,       2,   9,     10,   5,     6,         7
	// 0,   1,       3,   10,    10,   5,     6,         7
	// 0,   1,       4,   11,    10,   5,     6,         7
	// 0,   1,       5,   12,    10,   5,     6,         7
	// 0,   1,       6,   13,    10,   5,     6,         7
	// 0,   1,       7,   14,    10,   5,     6,         7
	// 1,   2,       8,   14,    10,   5,     6,         7
	for rows.Next() {
		if err := rows.Scan(&wIDStr, &w.Number,
			&dIDStr, &weekDateStr,
			&rIDStr, &r.Name, &r.Portions, &r.ProbabilityWeight); err != nil {
			return nil, err
		}
		var err error
		w.ID, err = uuid.Parse(wIDStr)
		if err != nil {
			return nil, err
		}
		d.ID, err = uuid.Parse(dIDStr)
		if err != nil {
			return nil, err
		}
		r.ID, err = uuid.Parse(rIDStr)
		if err != nil {
			return nil, err
		}
		if lastWeek == -1 {
			lastWeek = w.Number
		}
		if lastWeek != w.Number {
			w.Days = days
			days = []*app.Day{}
			weeks = append(weeks, w)
			lastWeek = w.Number
		}
		d.Dinner = r
		days = append(days, d)
	}

	return weeks, nil
}
