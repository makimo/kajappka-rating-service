package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// App holds the core app settings and required dependency objects,
// most importantly the RatingRepository instance must be passed for
// the app to be able to retrieve/save Ratings in some kind of store.
//
// Host and Port must be passed to instruct App where to bind to.
//
// It should be used as follows:
//
//     app := App{
//	       Host: Host,
//	       Port: Port,
//	       Repository: <repository>,
//     }
//     if err := app.Initialize(ctx); err != nil {
//         app.Run()
//     }
//
// Most important configuration happens in Initialize method.
type App struct {
	Host       string
	Port       string
	Repository RatingRepository

	router *mux.Router

	requestUserKey string
}

// getRatings route endpoint returns averaged game ratings in the following
// format:
//
//     [
//         {
//             "game_id": "first_game_id",
//             "rating": 5
//         },
//         {
//             "game_id": "second_game_id",
//             "rating": 4.9
//         },
//         ...
//     ]
//
// If no games have been rated, empty list is returned. Error with code
// 500 is returned in any other case.
func (a *App) getRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := a.Repository.GetAvgRatings(r.Context())

	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	if len(*ratings) == 0 {
		json.NewEncoder(w).Encode(make([]Rating, 0))
		return
	}

	json.NewEncoder(w).Encode(ratings)
}

// getRating route endpoint returns the currently logged in user rating
// for the game specified via `/{id}`` request variable in the URL. The
// rating is returned in the following format:
//
//     {
//         "game_id": "selected_game_id",
//         "rating": 3
//     }
//
// If game has not been rated before, 0 is returned as a rating. Error
// with code 500 is returned in any other case.
func (a *App) getRating(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(a.requestUserKey).(User)
	params := mux.Vars(r)

	rating, err := a.Repository.GetRating(r.Context(), params["id"], user.ID)

	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&rating)
}

// putRating route endpoint updates the currently logged in user rating
// for the game specified via `/{id}`` request variable in the URL. The
// rating must be provided in the following format:
//
//     {
//         "rating": 3
//     }
//
// and must be an integer between 1 and 5. If game has not been rated before,
// the rating is overwritten. In case of validation error, 400 Bad Request
// is returned. Error with code 500 is returned in any other case.
func (a *App) putRating(w http.ResponseWriter, r *http.Request) {
	var rating Rating

	user := r.Context().Value(a.requestUserKey).(User)
	params := mux.Vars(r)

	_ = json.NewDecoder(r.Body).Decode(&rating)
	rating.GameID = params["id"]
	rating.UserID = user.ID

	if err := a.Repository.PutRating(r.Context(), rating); err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	a.getRating(w, r)
}

// Initialize method calls Initialize on the provided repository (and returns)
// the error if any, configures route handlers on the Mux router and connects
// logging, authentication and request type middlewares to the request pipeline.
func (a *App) Initialize(ctx context.Context) error {
	if err := a.Repository.Initialize(ctx); err != nil {
		log.Println("Cannot connect to database")

		return err
	}

	a.requestUserKey = "user"

	a.router = mux.NewRouter()

	a.router.HandleFunc("/", a.getRatings).Methods("GET")
	a.router.HandleFunc("/{id}", a.getRating).Methods("GET")
	a.router.HandleFunc("/{id}", a.putRating).Methods("PUT")

	a.router.Use(LogRequestsMiddleware)

	a.router.Use(AuthenticationMiddleware{
		VerifierURI:    VerifierURI,
		RequestUserKey: a.requestUserKey,
	}.Middleware)

	a.router.Use(ContentTypeMiddleware)

	return nil
}

// Run invokes ListenAndServe on the previously configured Mux router
// and starts the event loop.
func (a *App) Run() {
	http.ListenAndServe(a.Host+":"+a.Port, a.router)
}
