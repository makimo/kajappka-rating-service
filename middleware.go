package main

import (
	"context"
	"log"
	"net/http"
)

// ContentTypeMiddleware returns `application/json` in `Content-Type`
// header for all requests its bound onto.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

// LogRequestsMiddleware logs request URL for all requests its bound to.
func LogRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// AuthenticationMiddleware authenticates the request using the authentication
// service. VerifierURI in the struct should be initialized with the URL
// for the authentication service. The middleware then should be used as:
//
//     AuthenticationMiddleware{VerifierURI: "URI"}.Middleware
//
// with Mux.
type AuthenticationMiddleware struct {
	VerifierURI    string
	RequestUserKey string
}

// Middleware passes the token found in `Authorization` header
// to the AuthenticateUser method which returns the `User` instance or error.
// When authenticated properly, the returned `User` is saved in request
// `Context` under `RequestUserKey` so it can be retrieved in route handlers.
func (a AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		user, err := AuthenticateUser(a.VerifierURI, token)

		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), a.RequestUserKey, *user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
