package app

type NewUser struct {
	Name string `json:"name,omitempty"`
}

type User struct {
	Id int64 `json:"id,omitempty"`
	NewUser
}

type Recipe struct {
	Name  string  `json:"name,omitempty"`
	Items []*Item `json:"items,omitempty"`
}

type Item struct {
	Name  string `json:"name,omitempty"`
	Price string `json:"price,omitempty"`
}

type UserService interface {
	User(id int) (*User, error)
	Users() ([]*User, error)
	CreateUser(u *NewUser) (*User, error)
	DeleteUser(id int) error
}
type RecipeService interface {
	Recipe(id int) (*Recipe, error)
	Recipes() ([]*Recipe, error)
	CreateRecipe(u *Recipe) (*Recipe, error)
	DeleteRecipe(id int) error
}
type ItemService interface {
	Item(id int) (*Item, error)
	Items() ([]*Item, error)
	CreateItem(u *Item) (*Item, error)
	DeleteItem(id int) error
}
