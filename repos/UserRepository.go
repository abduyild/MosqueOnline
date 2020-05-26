package repos

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
		return db.Collection("days"), nil
	}
	return nil, nil

}
