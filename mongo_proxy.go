// mongo_proxy.go
// David L. Flanagan
// July 23, 2019

package main

import (
	"context"
	"log"
	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoProxy ... Object for interacting with mongodb server.
type MongoProxy struct {
	client  *mongo.Client // The driver client object
	queries chan string   // A channel for accepting queries from the rest of the system.
	done    chan bool     // A channel for letting others know we are done.
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
		make(chan string),
		make(chan bool, 1),
	}, nil
}

// StartProcessing ... Handle the processing of queries recieved through the proxy's query channel.
func (mp *MongoProxy) StartProcessing() {
	// Forever...
	for {
		// Get a command and find out if the channel has been closed.
		command, done := <-mp.queries

		// Process the command.
		log.Println(command)

		// If the channel has been closed...
		if done {
			// Flush the channel to make sure we don't miss anything.
			for i := range mp.queries {
				log.Println(i)
			}
			// Let the rest of the system we are done here.
			mp.done <- true
			return
		}
	}
}

// Shutdown ... Dispose of the mongo proxy object.
func (mp *MongoProxy) Shutdown() {
	// Close the queries channel so the system will stop accepting queries.
	close(mp.queries)
	// Disconnect from teh mongodb server.
	err := mp.client.Disconnect(context.TODO())

	if err != nil {
		log.Fatalln(err)
	}

	// Wait for the query processing routine to complete...
	<-mp.done
	log.Println("MongoProxy shutdown complete...")
}
