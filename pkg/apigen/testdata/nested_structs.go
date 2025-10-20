package test

// Address represents a physical address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zipCode"`
}

// Company represents a company with nested address
type Company struct {
	Name    string  `json:"name"`
	Address Address `json:"address"`
}

// ProcessCompany processes company data
func ProcessCompany(company Company) error {
	return nil
}