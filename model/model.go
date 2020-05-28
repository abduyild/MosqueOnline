package model

import "labix.org/v2/mgo/bson"

// Structure of a User-struct
type User struct {
	FirstName         string `json:"FirstName"`
	LastName          string `json:"LastName"`
	Email             string `json:"Email"`
	Phone             string `json:"Phone"`
	Attended          bool   `json:"Attended"`
	RegisteredPrayers []RegisteredPrayer
}

//practically search with dataBase.FindId(bson.M{"_id": bson.ObjectIdHex("56bdd27ecfa93bfe3d35047d")})
type RegisteredPrayer struct {
	ID            bson.ObjectId `json:"id" bson:"_id,omitempty"` // ID of mosque
	MosqueName    string
	MosqueAddress string
	Date          string
	PrayerName    string
	DateIndex     int
	PrayerIndex   int
}
