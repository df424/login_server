// mongo_proxy.go
// David L. Flanagan
// July 23, 2019

package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoProxy ... Object for interacting with mongodb server.
type MongoProxy struct {
	client *mongo.Client // The driver client object
}

// NewMongoProxy ... Create a new mongo proxy and connect to the server...
func NewMongoProxy() (MongoProxy, error) {
	// Setup the mongo options and connect to the server.
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		return MongoProxy{}, err
	}

	// Make sure we can actually talk to the server.
	err = mongoClient.Ping(context.TODO(), nil)

	if err != nil {
		return MongoProxy{}, err
	}

	// Okay we are good return the object the system will use to interact with the db.
	return MongoProxy{
		mongoClient,
	}, nil
}

// GetUser ... Gets the information of a user if it exists.
func (mp *MongoProxy) GetUser(email string) (User, error) {
	// This object will store the user data if we find it.
	user := User{}

	// Query the database...
	result := mp.client.Database("userdb").Collection("users").FindOne(context.TODO(), bson.D{{"email", email}})

	// If we successfully got a user with the email...
	if result.Err() != nil {
		return user, result.Err()
	}

	// Decode the bson into a usable struct.
	result.Decode(&user)

	return user, nil
}
