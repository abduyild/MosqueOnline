package model

import "labix.org/v2/mgo/bson"

// Structure of a Mosque
type Mosque struct {
	ID     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name   string
	PLZ    int
	Street string
	City   string
	Date   []Date
}
