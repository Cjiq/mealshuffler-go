package sqlite

import (
	"database/sql"

	"nrdev.se/mealshuffler/app"
)

type RecipeService struct {
	db *sql.DB
}

func NewRecipeService(db *sql.DB) *RecipeService {
	us := &RecipeService{db: db}
	us.CreateRecipeTable()
	return &RecipeService{db: db}
}

func (u *RecipeService) CreateRecipeTable() error {
	query := `CREATE TABLE IF NOT EXISTS recipe (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		probability_weight REAL NOT NULL,
		portions INTEGER NOT NULL
		user_id INTEGER NOT NULL
	);`
	if _, err := u.db.Exec(query); err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS recipes_items (
		recipe_id INTEGER,
		item_id INTEGER,
	};
	CREATE INDEX idx_recipe_item ON recipes_items (recipe_id, item_id);
	CREATE INDEX idx_item ON recipes_items (item_id);
	`
	if _, err := u.db.Exec(query); err != nil {
		return err
	}

	return nil
}

// Recipes returns all recipes that are not owned by a user
func (u *RecipeService) Recipes() ([]*app.Recipe, error) {
	rows, err := u.db.Query(`SELECT id, name 
	FROM recepis
	WHERE user_id = null
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recepis := make([]*app.Recipe, 0)
	for rows.Next() {
		var r app.Recipe
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		recepis = append(recepis, &r)
	}

	return recepis, nil
}

func (u *RecipeService) CreateRecipe(newRecipe *app.NewRecipe) (*app.Recipe, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(`INSERT INTO 
	recipes(
		name, probability_weight, portions, created_at, updated_at
	)
	VALUES(?, ?, ?, datetime('now'), datetime('now'))
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		newRecipe.Name,
		newRecipe.ProbabilityWeight,
		newRecipe.Portions,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	recipe := &app.Recipe{
		Entity: app.Entity{
			ID: id,
		},
		NewRecipe: app.NewRecipe{
			Name:              newRecipe.Name,
			Portions:          newRecipe.Portions,
			ProbabilityWeight: newRecipe.ProbabilityWeight,
		},
	}
	tx.Commit()

	return recipe, nil
}

func (u *RecipeService) Recipe(id int) (*app.Recipe, error) {
	var recepi app.Recipe
	if err := u.db.QueryRow("SELECT id, name FROM recepis WHERE id = ?", id).Scan(&recepi.ID, &recepi.Name); err != nil {
		return nil, err
	}
	return &recepi, nil
}

func (u *RecipeService) DeleteRecipe(id int) error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("DELETE FROM recepis WHERE id = ?")
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
