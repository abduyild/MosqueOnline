package model

// Structure of a User-struct
type User struct {
	Username  string `json:"username"`
	Group     string `json:"group"`
	Password  string `json:"password"`
}
