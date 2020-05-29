package model

// Structure of a Mosque
type Mosque struct {
	Name   string `bson:"Name"`
	PLZ    int    `bson:"PLZ"`
	Street string `bson:"Street"`
	City   string `bson:"City"`
	Date   []Date `bson:"Date"`
}
