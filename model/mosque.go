package model

// Structure of a Mosque
type Mosque struct {
	Name          string `bson:"Name"`
	PLZ           int    `bson:"PLZ"`
	Street        string `bson:"Street"`
	MaxCapM       int    `bson:"MaxCapM"`
	MaxCapW       int    `bson:"MaxCapW"`
	City          string `bson:"City"`
	Date          []Date `bson:"Date"`
	Active        bool   `bson:"Active"`
	Cuma          bool   `bson:"Cuma"`
	Bayram        bool   `bson:"Bayram"`
	MaxFutureDate int    `bson:"MaxFutureDate" default:"5"`
	Ads           []Ad   `bson:"Ads"`
}

type Ad struct {
	Path string `bson:"Path"`
	Link string `bson:"Link"`
}
