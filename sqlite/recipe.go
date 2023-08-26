package sqlite

import (
	"database/sql"

	"github.com/google/uuid"

	"nrdev.se/mealshuffler/app"
)

type RecipeService struct {
	db *sql.DB
}

func NewRecipeService(db *sql.DB) *RecipeService {
	us := &RecipeService{db: db}
	err := us.CreateRecipeTable()
	if err != nil {
		panic(err)
	}
	return &RecipeService{db: db}
}

func (u *RecipeService) CreateRecipeTable() error {
	query := `CREATE TABLE IF NOT EXISTS recipe (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		probability_weight REAL NOT NULL,
		portions INTEGER NOT NULL,
		user_id TEXT
	);`
	if _, err := u.db.Exec(query); err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS recipes_items (
		recipe_id TEXT,
		item_id TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_recipe_item ON recipes_items (recipe_id, item_id);
	CREATE INDEX IF NOT EXISTS idx_item ON recipes_items (item_id);
	`
	if _, err := u.db.Exec(query); err != nil {
		return err
	}

	return nil
}

// Recipes returns all recipes that are not owned by a user
func (u *RecipeService) Recipes() ([]*app.Recipe, error) {
	rows, err := u.db.Query(`SELECT id, name, probability_weight, portions
	FROM recipe
	WHERE user_id is null OR user_id = ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recepis := make([]*app.Recipe, 0)
	for rows.Next() {
		var r app.Recipe
		if err := rows.Scan(&r.ID, &r.Name, &r.ProbabilityWeight, &r.Portions); err != nil {
			return nil, err
		}
		recepis = append(recepis, &r)
	}

	return recepis, nil
}

func (u *RecipeService) CreateRecipe(newRecipe *app.NewRecipe) (*app.Recipe, error) {
	id := uuid.New()
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(`INSERT INTO 
	recipe(
		id, name, probability_weight, portions
	)
	VALUES(?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		id.String(),
		newRecipe.Name,
		newRecipe.ProbabilityWeight,
		newRecipe.Portions,
	)
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
