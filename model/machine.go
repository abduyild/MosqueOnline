package model

type Machine struct {
	ID_Machines int
	SolvedUser bool
	SolvedRoot bool
	UserFlag string
	RootFlag string
}

type Group struct {
	ID int
	Points int
	Machines []Machine
}
