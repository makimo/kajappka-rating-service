package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// User describes a logged in user data returned from
// authentication endpoint. The game rating service relies
// only on ID field, but all others are stored for completeness.
type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Nickname     string `json:"nickname"`
	ProfilePhoto string `json:"profile_photo"`
}

// AuthenticateUser tries to authenticate `token` using against
// an external authentication service by passing that token through
// `Authorization` HTTP header.
//
// 200 OK from the authentication service is treated as success, then
// user data is expected in response and is decoded into an `User`
// instance and returned.
//
// Any other response from authentication service or actual error
// in connection or retrieving the response is passed as an error
// to the caller.
func AuthenticateUser(verifierURI string, token string) (*User, error) {
	req, err := http.NewRequest("GET", verifierURI, nil)

	if err != nil {
		log.Println("Error reading authentication request:", err)

		return nil, err
	}

	req.Header.Set("Authorization", token)

	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error reading authentication response:", err)

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var user User

		err := json.NewDecoder(resp.Body).Decode(&user)

		if err != nil {
			log.Println("Error decoding authentication response JSON:", err)

			return nil, err
		}

		log.Println("Authentication passed with:", token, "for user", user.ID)

		return &user, nil
	}

	log.Println("Authentication failed with:", token)

	return nil, errors.New("Authentication failed")
}
