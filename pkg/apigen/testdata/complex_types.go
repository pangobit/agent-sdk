package test

// Contact represents contact information
type Contact struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// Person represents a person with contacts
type Person struct {
	Name     string    `json:"name"`
	Contacts []Contact `json:"contacts"`
}

// Team represents a team with members
type Team struct {
	Name    string            `json:"name"`
	Members map[string]Person `json:"members"`
}

// ProcessTeam processes team data
func ProcessTeam(team Team) error {
	return nil
}