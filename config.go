package main

import (
	"log"
	"os"
)

var (
	// Host to bind to, e.g. 127.0.0.1
	Host string = os.Getenv("HOST")
	// Port to bind to, defaults to 8000
	Port string = os.Getenv("PORT")
	// VerifierURI is an authentication endpoint URI, required
	VerifierURI string = os.Getenv("VERIFIER_URI")
	// MongoConnectionString is a MongoDB connection string, required
	MongoConnectionString string = os.Getenv("MONGO_CONNECTION_STRING")
	// MongoDbName is a MongoDB database name, required
	MongoDbName string = os.Getenv("MONGO_DB_NAME")
	// MongoCollectionName is a MongoDB collection name, required
	MongoCollectionName string = os.Getenv("MONGO_COLLECTION_NAME")
)

// Environment variables values are validated and defaulted, if needed,
// before anything else gets called.
func init() {
	if VerifierURI == "" {
		log.Fatal("No VERIFIER_URI specified.")
	}

	if MongoConnectionString == "" {
		log.Fatal("No MONGO_CONNECTION_STRING specified.")
	}

	if MongoDbName == "" {
		log.Fatal("No MONGO_DB_NAME specified.")
	}

	if MongoCollectionName == "" {
		log.Fatal("No MONGO_COLLECTION_NAME specified.")
	}

	if Port == "" {
		Port = "8000"
	}
}
