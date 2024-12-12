package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Cprakhar/rss-aggregator/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	newUser, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: params.Name,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create new user: %v", err))
		return
	}

	respondJSONdata(w, 201, newUser)
}

func (apiCfg *apiConfig) handleGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondJSONdata(w, 200, user)
}