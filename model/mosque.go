package model

// Structure of a Group-struct, ID is GroupID, Points is the Points of the Group and Machines[] is an Slice (array) consisting of other Machine structures
type Mosque struct {
	Name string
	Capacity int
	PLZ  int
	Street string
	City string
	Users []User
}
