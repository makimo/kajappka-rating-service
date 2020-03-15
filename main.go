package main

import (
	"context"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app := App{
		Host: Host,
		Port: Port,
		Repository: &MongoRatingRepository{
			ConnectionString: MongoConnectionString,
			DatabaseName:     MongoDbName,
			CollectionName:   MongoCollectionName,
		},
	}

	if err := app.Initialize(ctx); err != nil {
		log.Fatal("Unrecoverable error, quitting")
	}

	app.Run()
}
