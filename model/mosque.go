package model

// Structure of a Mosque
type Mosque struct {
	Name    string `bson:"Name"`
	PLZ     int    `bson:"PLZ"`
	Street  string `bson:"Street"`
	MaxCapM int    `bson:"MaxCapM"`
	MaxCapW int    `bson:"MaxCapW"`
	City    string `bson:"City"`
	Date    []Date `bson:"Date"`
	Active  bool   `bson:"Active"`
}
