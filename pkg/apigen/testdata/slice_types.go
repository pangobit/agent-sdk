package test

// User represents a user
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProcessUsers processes a list of users
func ProcessUsers(users []User) error {
	return nil
}

// ProcessIDs processes a list of IDs
func ProcessIDs(ids []int) error {
	return nil
}