package model

type Admin struct {
	Name     string `bson:"Name"`
	Email    string `bson:"Email"`
	Password string `bson:"Password"`
	Admin    bool   `bson:"Admin"` // If overall admin, set true, else (mosque admin) set false
}
