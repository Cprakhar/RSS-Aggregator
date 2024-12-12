package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Cprakhar/rss-aggregator/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {

	type parameters struct {
		Feed_ID uuid.UUID `json:"feed_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	feedFollow, err := apiCfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
		FeedID: params.Feed_ID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't follow the feed: %v", err))
		return
	}

	respondJSONdata(w, 201, feedFollow)
}

func (apiCfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := apiCfg.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get the followed feeds: %v", err))
		return
	}
	respondJSONdata(w, 201, feedFollows)
}

func (apiCfg *apiConfig) handleDeleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	feedId, err := uuid.Parse(chi.URLParam(r, "feed_id"))
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't parse UUID: %v", err))
		return
	}
	err = apiCfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedId,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't delete the followed feed: %v", err))
		return
	}

	respondJSONdata(w, 200, struct{}{})
}
