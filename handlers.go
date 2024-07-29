package main

import (
	"errors"
	"memento/models"
	"memento/status"
	"memento/utils"
	"net/http"
	"strings"
)

func GetReviewsByAuthorID(w http.ResponseWriter, r *http.Request) {
	// Parse the user ID from the URL query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		status.ErrBadRequest(errors.New("Missing user_id parameter"))
		return
	}

	// Get the reviews by user ID
	reviews, movieNames, err := store.GetReviewsByUser(userID)
	if err != nil {
		status.ErrInternal(err)
		return
	}

	response := []*models.GetReviewsByUserResponse{}

	movieMap := make(map[string][]models.ReviewResponse)
	for i, review := range reviews {
		movieMap[movieNames[i]] = append(movieMap[movieNames[i]], models.ReviewResponse{
			AuthorID: review.AuthorID,
			Score:    review.Score,
			Comment:  review.Comment,
		})
	}

	for movieName, reviews := range movieMap {
		response = append(response, &models.GetReviewsByUserResponse{
			MovieName: movieName,
			Reviews:   reviews,
		})
	}

	utils.JsonEncode(w, http.StatusOK, response)
}

func GetAllMovies(w http.ResponseWriter, r *http.Request) {
	movies, averages, err := store.GetMovies()
	if err != nil {
		status.ErrInternal(err)
		return
	}

	response := []models.MovieResponse{}
	for i, movieName := range movies {
		response = append(response, models.MovieResponse{
			MovieName:    movieName,
			AverageScore: averages[i],
		})
	}

	utils.JsonEncode(w, http.StatusOK, response)
}

func GetReviewsByMovieName(w http.ResponseWriter, r *http.Request) {
	// Parse the movie name from the URL query parameters
	movieName := r.URL.Query().Get("movie_name")
	if movieName == "" {
		status.ErrBadRequest(errors.New("Missing movie_name parameter"))
		return
	}

	// Get the reviews and average score for the movie
	reviews, avg, err := store.GetReviews(strings.TrimSpace(movieName))
	if err != nil {
		status.ErrInternal(err)
		return
	}

	// Create the response
	reviewResponses := []models.ReviewResponse{}
	for _, review := range reviews {
		reviewResponses = append(reviewResponses, models.ReviewResponse{
			AuthorID: review.AuthorID,
			Score:    review.Score,
			Comment:  review.Comment,
		})
	}

	response := &models.MovieResponse{
		MovieName:    movieName,
		AverageScore: avg,
		Reviews:      reviewResponses,
	}

	utils.JsonEncode(w, http.StatusOK, response)
}

func AddReview(w http.ResponseWriter, r *http.Request) {
	// if req, err := utils.JsonDecode[*models.AddReviewRequest](r); err != nil {
	// 	status.ErrBadRequest(err)
	// 	return
	// }
}

// query parameter
func DeleteReview(w http.ResponseWriter, r *http.Request) {}

// query parameter
func ExamineMovie(w http.ResponseWriter, r *http.Request) {}

func RecommendMovie(w http.ResponseWriter, r *http.Request) {}
