package repos

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDBCollection(i int) (*mongo.Collection, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
	    return nil, err
	}
    // Check the connection
	err = client.Ping(context.TODO(), nil)
	var collection *mongo.Collection
	if err != nil {
	    return nil, err
	}
	if i == 0 {
	    collection = client.Database("PPT").Collection("users")
	} else if i == 1 {
	    collection = client.Database("PPT").Collection("groups")
	} else if i == 2 {
	    collection = client.Database("PPT").Collection("machines")
	}
    return collection, nil
	
}
