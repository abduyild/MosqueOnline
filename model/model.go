package model

// Structure of a User-struct
type User struct {
	Sex               string             `default:"Men" bson:"Sex" json:"Sex"`
	FirstName         string             `bson:"FirstName" json:"FirstName"`
	LastName          string             `bson:"LastName" json:"LastName"`
	Email             string             `bson:"Email" json:"Email"`
	Phone             string             `bson:"Phone" json:"Phone"`
	Attended          bool               `bson:"Attended" json:"Attended"`
	RegisteredPrayers []RegisteredPrayer `bson:"RegisteredPrayers" json:"RegisteredPrayers"`
}

type RegisteredPrayer struct {
	MosqueName    string `bson:"MosqueName"`
	MosqueAddress string `bson:"MosqueAddress"`
	Date          string `bson:"Date"`
	PrayerName    string `bson:"PrayerName"`
	DateIndex     int    `bson:"DateIndex"`
	PrayerIndex   int    `bson:"PrayerIndex"`
}
