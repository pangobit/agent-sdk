// Package api provides comprehensive examples of all supported API generation patterns
package api

import "time"

// BasicTypes demonstrates basic Go types
type BasicTypes struct{}

// HandleBasicTypes processes basic Go types
// @param message A simple string message
// @param count An integer count
// @param enabled A boolean flag
// @param rate A floating point rate
func (bt *BasicTypes) HandleBasicTypes(message string, count int, enabled bool, rate float64) error {
	return nil
}

// NestedStructs demonstrates nested struct types
type Address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	State   string `json:"state" validate:"required"`
	ZipCode string `json:"zipCode" validate:"required"`
}

type Company struct {
	Name    string  `json:"name" validate:"required"`
	Address Address `json:"address"`
}

// ProcessCompany processes a company with nested address
func (bt *BasicTypes) ProcessCompany(company Company) error {
	return nil
}

// SliceTypes demonstrates slice/array types
type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	IsActive bool     `json:"isActive"`
	Tags     []string `json:"tags,omitempty"`
}

// ProcessUsers processes a slice of users
func (bt *BasicTypes) ProcessUsers(users []User) error {
	return nil
}

// ProcessIDs processes a slice of integers
func (bt *BasicTypes) ProcessIDs(ids []int) error {
	return nil
}

// MapTypes demonstrates map types
type Profile struct {
	Bio     string   `json:"bio"`
	Age     int      `json:"age"`
	Active  bool     `json:"active"`
	Tags    []string `json:"tags"`
}

// ProcessProfiles processes a map of user profiles
func (bt *BasicTypes) ProcessProfiles(profiles map[string]Profile) error {
	return nil
}

// ProcessData processes a simple string map
func (bt *BasicTypes) ProcessData(data map[string]string) error {
	return nil
}

// ComplexTypes demonstrates complex nested structures
type Project struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Owner       User                `json:"owner"`
	Members     []User              `json:"members"`
	Metadata    map[string]string  `json:"metadata"`
}

type Team struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Projects    []Project          `json:"projects"`
	Lead        User               `json:"lead"`
	Settings    map[string]string  `json:"settings"`
}

// ProcessTeam processes a complex nested team structure
func (bt *BasicTypes) ProcessTeam(team Team) error {
	return nil
}

// PointerTypes demonstrates pointer types
type Config struct {
	DatabaseURL string `json:"databaseUrl"`
	Timeout     int    `json:"timeout"`
}

// UpdateConfig updates configuration with pointer types
func (bt *BasicTypes) UpdateConfig(config *Config) error {
	return nil
}

// SelectorTypes demonstrates types from other packages

// ScheduleEvent schedules an event with time types
func (bt *BasicTypes) ScheduleEvent(name string, scheduledAt time.Time) error {
	return nil
}