package model
// Structure of a Machine-struct
type Machine struct {
	ID_Machines int
	SolvedUser  bool
	SolvedRoot  bool
	UserFlag    string
	RootFlag    string
}

// Structure of a Group-struct, ID is GroupID, Points is the Points of the Group and Machines[] is an Slice (array) consisting of other Machine structures
type Group struct {
	ID       int
	Points   int
	Machines []Machine
}
