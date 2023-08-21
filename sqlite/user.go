package sqlite

import (
	"database/sql"

	"nrdev.se/mealshuffler/app"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	us := &UserService{db: db}
	us.CreateUserTable()
	return &UserService{db: db}
}

func (u *UserService) CreateUserTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
);`
	if _, err := u.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (u *UserService) Users() ([]*app.User, error) {
	rows, err := u.db.Query("SELECT id, name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*app.User, 0)
	for rows.Next() {
		var u app.User
		if err := rows.Scan(&u.Id, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, nil
}

func (u *UserService) CreateUser(newUser *app.NewUser) (*app.User, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO users(name) VALUES(?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(newUser.Name)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := &app.User{
		Id: id,
		NewUser: app.NewUser{
			Name: newUser.Name,
		},
	}
	tx.Commit()

	return user, nil
}

func (u *UserService) User(id int) (*app.User, error) {
	var user app.User
	if err := u.db.QueryRow("SELECT id, name FROM users WHERE id = ?", id).Scan(&user.Id, &user.Name); err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserService) DeleteUser(id int) error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("DELETE FROM users WHERE id = ?")
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
