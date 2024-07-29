package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func GetReviewsByAuthorID(w http.ResponseWriter, r *http.Request) {
	// Parse the user ID from the URL query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	// Get the reviews by user ID
	reviews, movieNames, err := store.GetReviewsByUser(userID)
	if err != nil {
		http.Error(w, "Error fetching reviews: "+err.Error(), http.StatusInternalServerError)
		return
	}

	movieMap := make(map[string][]ReviewResponse)
	for i, review := range reviews {
		movieMap[movieNames[i]] = append(movieMap[movieNames[i]], ReviewResponse{
			AuthorID: review.AuthorID,
			Score:    review.Score,
			Comment:  review.Comment,
		})
	}
	for movieName, reviews := range movieMap {
		response = append(response, struct {
			MovieName string           `json:"movie_name"`
			Reviews   []ReviewResponse `json:"reviews"`
		}{
			MovieName: movieName,
			Reviews:   reviews,
		})
	}

	// Write the JSON response using json.Encoder
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetAllMovies(w http.ResponseWriter, r *http.Request) {
	// Get all movies and their average scores
	movies, averages, err := store.GetMovies()
	if err != nil {
		http.Error(w, "Error fetching movies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response structure
	type MovieResponse struct {
		MovieName    string  `json:"movie_name"`
		AverageScore float64 `json:"average_score"`
	}

	var response []MovieResponse
	for i, movieName := range movies {
		response = append(response, MovieResponse{
			MovieName:    movieName,
			AverageScore: averages[i],
		})
	}

	// Write the JSON response using json.Encoder
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetReviewsByMovieName(w http.ResponseWriter, r *http.Request) {
	// Parse the movie name from the URL query parameters
	movieName := r.URL.Query().Get("movie_name")
	if movieName == "" {
		http.Error(w, "Missing movie_name parameter", http.StatusBadRequest)
		return
	}

	// Get the reviews and average score for the movie
	reviews, avg, err := store.GetReviews(strings.TrimSpace(movieName))
	if err != nil {
		http.Error(w, "Error fetching reviews: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the response
	var reviewResponses []ReviewResponse
	for _, review := range reviews {
		reviewResponses = append(reviewResponses, ReviewResponse{
			AuthorID: review.AuthorID,
			Score:    review.Score,
			Comment:  review.Comment,
		})
	}

	response := MovieResponse{
		MovieName:    movieName,
		AverageScore: avg,
		Reviews:      reviewResponses,
	}

	// Write the JSON response using json.Encoder
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
