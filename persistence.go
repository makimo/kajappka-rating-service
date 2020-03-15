package main

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Rating represents a single game rating. UserID contains the user
// identifier, never encoded into JSON.
//
// Rating is valid if it falls between 1 and 5 and is an integer.
type Rating struct {
	UserID string `json:"-" bson:"user_id"`
	GameID string `json:"game_id" bson:"game_id"`
	Rating int    `json:"rating" bson:"rating"`
}

func (r Rating) valid() bool {
	return r.Rating >= 1 && r.Rating <= 5
}

// RatingRepository defines an interface allowing read/write access to the
// underlying store and can be implemented by multiple providers.
type RatingRepository interface {
	Initialize(ctx context.Context) error
	GetAvgRatings(ctx context.Context) (*[]Rating, error)
	GetRating(ctx context.Context, gameID string, userID string) (*Rating, error)
	PutRating(ctx context.Context, rating Rating) error
}

// MongoRatingRepository implements the RatingRepository using MongoDB
// database.
type MongoRatingRepository struct {
	ConnectionString string
	DatabaseName     string
	CollectionName   string

	collection *mongo.Collection
}

// Initialize opens connection to MongoDB database, pings it to verify
// connectivity and selects the database and collection for use. Collection
// pointer is saved in the repository object for use by individual store
// methods. Error is returned in any abnormal situation.
func (r *MongoRatingRepository) Initialize(ctx context.Context) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.ConnectionString))

	if err != nil {
		return err
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		return err
	}

	r.collection = client.Database(r.DatabaseName).Collection(r.CollectionName)

	log.Println("Successfully connected to", r.ConnectionString)

	return nil
}

// GetAvgRatings returns `*[]Rating` slice containing averaged ratings for
// all games in descending order, sorted by average rating.
func (r *MongoRatingRepository) GetAvgRatings(ctx context.Context) (*[]Rating, error) {
	var ratings []Rating

	pipeline := []bson.M{
		bson.M{
			"$group": bson.M{
				"_id": "$game_id",
				"game_id": bson.M{
					"$first": "$game_id",
				},
				"rating": bson.M{
					"$avg": "$rating",
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"rating": -1,
			},
		},
	}

	cur, err := r.collection.Aggregate(ctx, pipeline)

	if err != nil {
		log.Println("Error retrieving ratings: ", err)
		return nil, err
	}

	err = cur.All(ctx, &ratings)

	if err != nil {
		log.Println("Error retrieving ratings: ", err)
		return nil, err
	}

	return &ratings, nil
}

// GetRating returns `Rating` object for a given `gameID` and `userID`
func (r *MongoRatingRepository) GetRating(
	ctx context.Context,
	gameID string,
	userID string,
) (*Rating, error) {
	var rating Rating

	filter := bson.M{
		"game_id": gameID,
		"user_id": userID,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&rating)

	if err != nil && err != mongo.ErrNoDocuments {
		log.Println("Error retrieving rating: ", err)

		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return &Rating{
			GameID: gameID,
			Rating: 0,
		}, nil
	}

	return &rating, nil
}

// PutRating updates store with a new rating based on a given `Rating` object.
// Game identifier, user identifier and rating itself are taken from the object.
func (r *MongoRatingRepository) PutRating(ctx context.Context, rating Rating) error {
	if !rating.valid() {
		log.Println("Invalid rating update: ", rating)

		return errors.New("Invalid rating update")
	}

	filter := bson.M{
		"game_id": rating.GameID,
		"user_id": rating.UserID,
	}

	options := options.Replace().SetUpsert(true)

	_, err := r.collection.ReplaceOne(ctx, filter, &rating, options)

	if err != nil {
		log.Println("Error updating rating: ", err)

		return err
	}

	return nil
}
