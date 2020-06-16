package repos

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"pi-software/model"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	keyFile = "aes.key"
)

var IV = []byte("1234567812345678")

func GetDBCollection(i int) (*mongo.Collection, error) {
	// Define Address of Database
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
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
	}
	return nil, nil
}

func readKey(filename string) ([]byte, error) {
	key, err := ioutil.ReadFile(filename)
	if err != nil {
		return key, err
	}
	block, _ := pem.Decode(key)
	return block.Bytes, nil
}

func createKey() []byte {
	genkey := make([]byte, 16)
	_, err := rand.Read(genkey)
	if err != nil {
		log.Fatalf("Failed to read new random key: %s", err)
	}
	return genkey
}

func saveKey(filename string, key []byte) {
	block := &pem.Block{
		Type:  "AES KEY",
		Bytes: key,
	}
	err := ioutil.WriteFile(filename, pem.EncodeToMemory(block), 0644)
	if err != nil {
		log.Fatalf("Failed in saving key to %s: %s", filename, err)
	}
}

func aesKey() []byte {
	file := fmt.Sprintf(keyFile)
	key, err := readKey(file)
	if err != nil {
		log.Println("Creating a new AES key")
		key = createKey()
		saveKey(file, key)
	}
	return key
}

func createCipher() cipher.Block {
	c, err := aes.NewCipher(aesKey())
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
	return encode(bytes)
}

func Decrypt(cipherText string) string {
	bytes := decode(cipherText)
	blockCipher := createCipher()
	stream := cipher.NewCTR(blockCipher, IV)
	stream.XORKeyStream(bytes, bytes)
	return string(bytes)
}

func encode(input []byte) string {
	return string(b64.StdEncoding.EncodeToString(input))
}

func decode(input string) []byte {
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
			for j, prayer := range date.Prayer {
				for k := range prayer.Users {
					newMosque.Date[i].Prayer[j].Users[k] = emptyUser
				}
			}
		}
		collection.ReplaceOne(context.TODO(), bson.M{"Name": mosq.Name}, newMosque)
	}
}
