package model

// Structure of a User-struct
type User struct {
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Email     string `json:"Email"`
	Phone     string `json:"Phone"`
	Attended  bool   `json:"Attended"`
}
