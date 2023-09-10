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
	rs := &RecipeService{db: db}
	err := rs.CreateRecipeTable()
	if err != nil {
		panic(err)
	}
	return rs
}

func (r *RecipeService) CreateRecipeTable() error {
	query := `CREATE TABLE IF NOT EXISTS recipe (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		probability_weight REAL NOT NULL,
		portions INTEGER NOT NULL,
		left_over_compliance INTEGER NOT NULL,
		url TEXT,
		user_id TEXT
	);`
	if _, err := r.db.Exec(query); err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS recipes_items (
		recipe_id TEXT,
		item_id TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_recipe_item ON recipes_items (recipe_id, item_id);
	CREATE INDEX IF NOT EXISTS idx_item ON recipes_items (item_id);
	`
	if _, err := r.db.Exec(query); err != nil {
		return err
	}

	return nil
}

// Recipes returns all recipes that are not owned by a user
func (r *RecipeService) Recipes() ([]*app.Recipe, error) {
	return r.getRecipes()
}

func (r *RecipeService) UserRecipes(userID string) ([]*app.Recipe, error) {
	return r.getRecipes(userID)
}

func (r *RecipeService) getRecipes(userID ...string) ([]*app.Recipe, error) {
	var uID string
	if len(userID) > 0 {
		uID = userID[0]
	}

	rows, err := r.db.Query(`SELECT 
		id, name, probability_weight, portions, left_over_compliance, url
	FROM recipe
	WHERE user_id = ? or user_id is null or user_id = ''
	`, uID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recepis := make([]*app.Recipe, 0)
	for rows.Next() {
		var r app.Recipe
		var leftOverCompliance sql.NullInt64
		var url sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &r.ProbabilityWeight, &r.Portions, &leftOverCompliance, &url); err != nil {
			return nil, err
		}
		if leftOverCompliance.Valid && leftOverCompliance.Int64 > 0 {
			r.LeftOverCompliance = true
		}
		if url.Valid {
			r.URL = url.String
		}
		recepis = append(recepis, &r)
	}

	return recepis, nil

}

func (r *RecipeService) CreateRecipe(newRecipe *app.NewRecipe, userID string) (*app.Recipe, error) {
	id := uuid.New()
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(`INSERT INTO 
	recipe(
		id, name, probability_weight, portions, left_over_compliance, url, user_id
	)
	VALUES(?, ?, ?, ?, ?, ?, ?)
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
		newRecipe.LeftOverCompliance,
		newRecipe.URL,
		userID,
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

func (r *RecipeService) Recipe(id int) (*app.Recipe, error) {
	var recepi app.Recipe
	if err := r.db.QueryRow("SELECT id, name FROM recepis WHERE id = ?", id).Scan(&recepi.ID, &recepi.Name); err != nil {
		return nil, err
	}
	return &recepi, nil
}

func (r *RecipeService) DeleteRecipe(id int) error {
	tx, err := r.db.Begin()
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

func (rs *RecipeService) DeleteAllRecipes() error {
	tx, err := rs.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("DELETE FROM recipe")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (rs *RecipeService) UpdateRecipe(recipe *app.Recipe) (*app.Recipe, error) {
	tx, err := rs.db.Begin()
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("UPDATE recipe SET name = ?, probability_weight = ?, portions = ? WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(recipe.Name, recipe.ProbabilityWeight, recipe.Portions, recipe.ID.String()); err != nil {
		return nil, err
	}
	tx.Commit()
	return recipe, nil
}
