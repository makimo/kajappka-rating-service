package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type key int

const requestUserKey key = 0

var host string = os.Getenv("HOST")
var port string = os.Getenv("PORT")
var verifierURI string = os.Getenv("VERIFIER_URI")
var mongoConnectionString string = os.Getenv("MONGO_CONNECTION_STRING")
var mongoDbName string = os.Getenv("MONGO_DB_NAME")
var mongoCollectionName string = os.Getenv("MONGO_COLLECTION_NAME")

var collection *mongo.Collection

// User describes a logged in user
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Nickname     string `json:"nickname"`
	ProfilePhoto string `json:"profile_photo"`
}

// Rating represents a single game rating
type Rating struct {
	UserID string  `json:"-" bson:"user_id"`
	GameID string  `json:"game_id" bson:"game_id"`
	Rating float64 `json:"rating" bson:"rating"`
}

func getRatings(w http.ResponseWriter, r *http.Request) {
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

	cur, err := collection.Aggregate(r.Context(), pipeline)

	if err != nil {
		log.Println("Error retrieving ratings: ", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	err = cur.All(r.Context(), &ratings)

	if err != nil {
		log.Println("Error retrieving ratings: ", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	log.Println("getRatings:", cur)
	json.NewEncoder(w).Encode(ratings)
}

func getRating(w http.ResponseWriter, r *http.Request) {
	var rating Rating

	user := r.Context().Value(requestUserKey).(User)

	params := mux.Vars(r)

	filter := bson.M{
		"game_id": params["id"],
		"user_id": user.ID,
	}

	err := collection.FindOne(r.Context(), filter).Decode(&rating)

	if err != nil {
		log.Println("Error retrieving rating: ", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(&rating)
}

func putRating(w http.ResponseWriter, r *http.Request) {
	var rating Rating

	user := r.Context().Value(requestUserKey).(User)

	params := mux.Vars(r)

	filter := bson.M{
		"game_id": params["id"],
		"user_id": user.ID,
	}

	_ = json.NewDecoder(r.Body).Decode(&rating)
	rating.GameID = params["id"]
	rating.UserID = user.ID

	options := options.Replace().SetUpsert(true)

	_, err := collection.ReplaceOne(r.Context(), filter, &rating, options)

	if err != nil {
		log.Fatal("Error inserting rating: ", err)
	}

	getRating(w, r)
}

func logRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		req, err := http.NewRequest("GET", verifierURI, nil)

		if err != nil {
			log.Println("Error reading authentication request:", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		req.Header.Set("Authorization", token)

		client := &http.Client{Timeout: time.Second * 5}
		resp, err := client.Do(req)

		if err != nil {
			log.Println("Error reading authentication response:", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var user User

			err := json.NewDecoder(resp.Body).Decode(&user)

			if err != nil {
				log.Println("Error decoding authentication response JSON:", err)

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			log.Println("Authentication passed with:", token, "for user", user.ID)

			ctx := context.WithValue(r.Context(), requestUserKey, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			log.Println("Authentication failed with:", token)

			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

func main() {
	if verifierURI == "" {
		log.Fatal("No VERIFIER_URI specified.")
	}

	if port == "" {
		port = "8000"
	}

	router := mux.NewRouter()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnectionString))

	if err != nil {
		log.Fatal("MongoDB connection failed: ", err)
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Fatal("MongoDB connection failed: ", err)
	}

	collection = client.Database(mongoDbName).Collection(mongoCollectionName)

	router.HandleFunc("/", getRatings).Methods("GET")
	router.HandleFunc("/{id}", getRating).Methods("GET")
	router.HandleFunc("/{id}", putRating).Methods("PUT")

	router.Use(logRequestsMiddleware)
	router.Use(authenticationMiddleware)
	router.Use(contentTypeMiddleware)

	http.ListenAndServe(host+":"+port, router)
}
