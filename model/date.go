package model

import "time"

type PrayerName int

var bayramList []string

const (
	Sabah  PrayerName = iota //1
	Ögle                     //2
	Ikindi                   //3
	Aksam                    //4
	Yatsi                    //5
	Cuma                     //6
	Bayram                   //7
)

type Date struct {
	Date   time.Time `bson:"Date"`
	Prayer []Prayer  `bson:"Prayer"`
}

type Prayer struct {
	Name          PrayerName `bson:"Name"`
	Available     bool       `bson:"Available"`
	CapacityMen   int        `bson:"CapacityMen"`
	CapacityWomen int        `bson:"CapacityWomen"`
	Users         []User     `bson:"Users"`
}

func GetBayrams() []string { // just add new bayrams here
	bayramList = []string{"2020-07-30"}
	return bayramList
}
