package model

// Structure of a User-struct
type User struct {
	// TODO:  Geschlecht mit einbringen
	FirstName         string `json:"FirstName"`
	LastName          string `json:"LastName"`
	Email             string `json:"Email"`
	Phone             string `json:"Phone"`
	Attended          bool   `json:"Attended"`
	RegisteredPrayers []RegisteredPrayer
}

type RegisteredPrayer struct {
	MosqueName    string
	MosqueAddress string
	Date          string
	PrayerName    string
	DateIndex     int
	PrayerIndex   int
}
