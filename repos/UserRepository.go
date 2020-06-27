package repos

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	b64 "encoding/base64"
	"errors"
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
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetAuth(options.Credential{AuthSource: "admin", Username: "mosquo", Password: "-MosqueOnline202066+"})
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
		// Get The Mosques Table with the entries of the Mosques
		return db.Collection("mosques"), nil
	} else if i == 2 {
		// Get The Mosques Table with the entries of the Mosques
		return db.Collection("admins"), nil
	} else if i == 3 {
		return db.Collection("eids"), nil
	}
	return nil, errors.New("Veribankada hata olusdu, birdaha deneyin | Datenbankfehler, versuchen Sie es erneut")
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
	gocron.Every(2).Weeks().Do(overwrite)
	gocron.Start()
}

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

			/* previous version:
			for j, prayer := range date.Prayer {
				for k := range prayer.Users {
					newMosque.Date[i].Prayer[j].Users[k] = emptyUser
				}
			}
			*/
		}
		collection.ReplaceOne(context.TODO(), bson.M{"Name": mosq.Name}, newMosque)
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
