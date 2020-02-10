package model

type User struct {
	Username string `json:"username"`
	Group    string `json:"group"`
	Password string `json:"password"`
}
