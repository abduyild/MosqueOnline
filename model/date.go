package model

import "time"

type prayerName int

const (
	Sabah  prayerName = iota //1
	Ögle                     //2
	Ikindi                   //3
	Aksam                    //4
	Yatsi                    //5
	Cuma                     //6
	Bayram                   //7
)

type Date struct {
	Date   time.Time
	Prayer []Prayer
}

type Prayer struct {
	Name prayerName
	//TODO capacity to men and women seperately, simpy make array, [0] for men, [1] women
	Capacity int
	Users    []User
}
