package repos

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	b64 "encoding/base64"
	"log"
	"pi-software/model"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const keyFile = "aes.key"

var key = []byte{11, 108, 111, 57, 116, 83, 193, 127, 59, 57, 245, 188, 171, 59, 187, 101}

var IV = []byte("1234567812345678")

func GetDBCollection(i int) (*mongo.Collection, error) {
	// Define Address of Database
	clientOptions := options.Client().ApplyURI("mongodb://0.0.0.0:27017").SetAuth(options.Credential{AuthSource: "admin", Username: "mosquo", Password: "-MosqueOnline202066+"})
	// Try to connect to Database, save error if one is thrown
	client, err := mongo.Connect(context.TODO(), clientOptions)
	// If there was an error connecting to the DB (DB not running, wrong URI, ...) return the error
	if err != nil {
		return nil, err
	}

	// Check if connection could be established to running DB

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	// Define the name of the Database as PPT, change this if you want to name your DB otherwise
	db := client.Database("MosqueOnline")
	// Working with int for extensibility, you can just add another else if and check for another value if you want to add another table
	// Get the Users Table
	if i == 0 {
		return db.Collection("users"), nil
	} else if i == 1 {
		return db.Collection("mosques"), nil
	} else if i == 2 {
		return db.Collection("admins"), nil
	} else if i == 3 {
		return db.Collection("eids"), nil
	}
	return nil, nil
}

func createCipher() cipher.Block {
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Failed to create the AES cipher: %s", err)
	}
	return c
}

func Encrypt(plainText string) string {
	bytes := []byte(plainText)
	blockCipher := createCipher()
	stream := cipher.NewCTR(blockCipher, IV)
	stream.XORKeyStream(bytes, bytes)
	return Encode(bytes)
}

func Decrypt(cipherText string) string {
	bytes := Decode(cipherText)
	blockCipher := createCipher()
	stream := cipher.NewCTR(blockCipher, IV)
	stream.XORKeyStream(bytes, bytes)
	return string(bytes)
}

func Encode(input []byte) string {
	return string(b64.StdEncoding.EncodeToString(input))
}

func Decode(input string) []byte {
	dcd, err := b64.StdEncoding.DecodeString(input)
	if err != nil {
		return []byte{}
	}
	return dcd
}

func StartCronjob() {
	gocron.Every(2).Weeks().At("03:00").Do(overwrite)
	gocron.Start()
}

// instead of deleting data delete elements, dont forget to iterate through all users and adjust index etc.
func overwrite() {
	collection, err := GetDBCollection(1)
	if err != nil {
		panic(err.Error())
	}
	var emptyUser model.User
	emptyUser.RegisteredPrayers = []model.RegisteredPrayer{}
	var newMosque model.Mosque
	var mosques []model.Mosque
	today := strings.Split(time.Now().String(), " ")[0]
	cur, _ := collection.Find(context.TODO(), bson.M{})
	for cur.Next(context.TODO()) {
		var mosque model.Mosque
		cur.Decode(&mosque)
		mosques = append(mosques, mosque)
	}
	for _, mosq := range mosques {
		newMosque = mosq
		for i, date := range mosq.Date {
			if today == strings.Split(date.Date.String(), " ")[0] {
				break
			}
			for j := range date.Prayer {
				newMosque.Date[i].Prayer[j].Users = []model.User{}
			}
		}
		collection.ReplaceOne(context.TODO(), bson.M{"Name": mosq.Name}, newMosque)
	}
	addDates()
}

func addDates() {
	collection, err := GetDBCollection(1)
	if err != nil {
		panic(err.Error())
	}
	cur, _ := collection.Find(context.TODO(), bson.M{})
	for cur.Next(context.TODO()) {
		var mosque model.Mosque
		cur.Decode(&mosque)

		var prayer model.Prayer
		var prayers = make([]model.Prayer, 7)
		prayer.CapacityMen = mosque.MaxCapM
		prayer.CapacityWomen = mosque.MaxCapW
		prayer.Users = []model.User{}
		cumaSet := mosque.Cuma
		bayramSet := mosque.Bayram
		for i := 1; i < 6; i++ {
			switch i {
			case 1:
				prayer.Available = mosque.Date[0].Prayer[0].Available
			case 2:
				prayer.Available = mosque.Date[0].Prayer[1].Available
			case 3:
				prayer.Available = mosque.Date[0].Prayer[2].Available
			case 4:
				prayer.Available = mosque.Date[0].Prayer[3].Available
			case 5:
				prayer.Available = mosque.Date[0].Prayer[4].Available
			}
			prayer.Name = model.PrayerName(i)
			prayers[i-1] = prayer
			prayer.Available = false
		}
		length := len(mosque.Date)
		for i := 1; i < 15; i++ {
			mosqueDate := mosque.Date[length-1].Date

			var date model.Date
			currentDate := mosqueDate.AddDate(0, 0, i).Format(time.RFC3339)
			weekday := mosqueDate.AddDate(0, 0, i).Weekday()
			if cumaSet && int(weekday) == 5 { // cuma
				prayers[5].Available = true
			}
			eids := GetEids()
			if bayramSet && containString(eids, strings.Split(currentDate, "T")[0]) {
				prayers[6].Available = true
			}
			date.Date, _ = time.Parse(time.RFC3339, currentDate)
			date.Prayer = prayers

			collection.UpdateOne(context.TODO(),
				bson.M{"Name": mosque.Name},
				bson.M{"$push": bson.M{"Date": date}})
			prayers[5].Available = false
			prayers[6].Available = false
		}
	}
}

type eidStruct struct {
	Date string `bson:"Date"`
}

func GetEids() []string {
	collection, _ := GetDBCollection(3)
	cur, _ := collection.Find(context.TODO(), bson.M{})
	eids := []string{}
	var eid eidStruct
	for cur.Next(context.TODO()) {
		cur.Decode(&eid)
		eids = append(eids, eid.Date)
	}
	return eids
}

func AddEid(input string) {
	collection, _ := GetDBCollection(3)
	eid := eidStruct{Date: input}
	collection.InsertOne(context.TODO(), eid)
}

func RemoveEid(input string) {
	collection, _ := GetDBCollection(3)
	collection.DeleteOne(context.TODO(), bson.M{"Date": input})
}

func containString(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}
