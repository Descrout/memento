package main

import (
	"errors"
	"fmt"
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
		status.ErrBadRequest(errors.New("missing user_id parameter"))
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
		status.ErrBadRequest(errors.New("missing movie_name parameter"))
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

	response := &models.MovieWithReviewsResponse{
		MovieName:    movieName,
		AverageScore: avg,
		Reviews:      reviewResponses,
	}

	utils.JsonEncode(w, http.StatusOK, response)
}

func AddReview(w http.ResponseWriter, r *http.Request) {

	req, err := utils.JsonDecode[*models.AddReviewRequest](r)
	if err != nil {
		status.ErrBadRequest(err)
		return
	}

	// Validate the request data
	if req.AuthorID == "" || req.MovieName == "" || req.Score < 1 || req.Score > 10 || req.Comment == "" {
		status.ErrBadRequest(errors.New("invalid request data"))
		return
	}

	// Create a new review
	review := &models.Review{
		AuthorID: req.AuthorID,
		Score:    req.Score,
		Comment:  req.Comment}

	// Add the review to the store
	if err := store.AddReview(req.MovieName, review); err != nil {
		status.ErrInternal(err).Render(w)
		return
	}

	// Respond with a success message
	status.StatusCreated("Review added successfully").Render(w)

}

// query parameter
func DeleteReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	movieName := r.URL.Query().Get("movie_name")

	// Validate the request data
	if userID == "" || movieName == "" {
		status.ErrBadRequest(errors.New("missing user_id or movie_name parameter"))
		return
	}

	// Delete the review from the store
	if err := store.DeleteReview(movieName, userID); err != nil {
		status.ErrInternal(err)
		return
	}

	// NOT SURE but Check if the movie still has reviews, if not, delete the movie
	movieDeleted := false
	count, err := store.GetReviewCount(movieName)
	if err == nil && count == 0 {
		if err := store.DeleteMovie(movieName); err == nil {
			status.ErrInternal(err).Render(w)
			return
		}
		movieDeleted = true
	}

	// Respond with a success message
	result := "Review deleted successfully."
	if movieDeleted {
		result += " Movie has been deleted because no reviews are left."
	}
	status.StatusOK(result).Render(w)
}

// query parameter
func ExamineMovie(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	movieName := r.URL.Query().Get("movie_name")
	personal := r.URL.Query().Get("personal")

	// Validate the request data
	if userID == "" || movieName == "" {
		status.ErrBadRequest(errors.New("missing user_id or movie_name parameter"))
		return
	}

	// Create the request text for ChatGPT
	var requestText strings.Builder
	requestText.WriteString("Sana film listesi ve onlara verdiğim puanları vereceğim. Bu puanlardan yola çıkarak sence " + movieName + " filmi hakkında ne düşünürüm? Sever miyim? İzlenir mi?\n\nListe:\n")

	if personal == "true" {
		reviews, names, err := store.GetReviewsByUser(userID)
		if err != nil {
			status.ErrInternal(err)
			return
		}
		for i, review := range reviews {
			requestText.WriteString(fmt.Sprintf("%s - Score: %.1f\n", names[i], review.Score))
		}
	} else {
		movies, averages, err := store.GetMovies()
		if err != nil {
			status.ErrInternal(err)
			return
		}
		for i, movie := range movies {
			requestText.WriteString(fmt.Sprintf("%s - Average Score: %.1f\n", movie, averages[i]))
		}
	}

	// Send the request to ChatGPT
	examination, err := utils.ChatGPTRequest(requestText.String())
	if err != nil {
		status.ErrInternal(err)
		return
	}

	status.StatusOK(examination).Render(w)
}

func RecommendMovie(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	personal := r.URL.Query().Get("personal")

	// Validate the request data
	if userID == "" {
		status.ErrBadRequest(errors.New("missing user_id parameter"))
		return
	}

	// Create the request text for ChatGPT
	var requestText strings.Builder
	requestText.WriteString("Sana film listesi ve onlara verdiğim puanları vereceğim. Bunlara göre bana beğenebileceğim 3 film öner:\n\nListe:\n")

	if personal == "true" {
		reviews, names, err := store.GetReviewsByUser(userID)
		if err != nil || len(reviews) == 0 {
			responseMessage := "You haven't reviewed any movies yet, so recommendations cannot be provided."
			if err != nil {
				responseMessage = fmt.Sprintf("AI Recommendation failed: %s", err.Error())
			}
			status.StatusOK(responseMessage)
			return
		}

		for i, review := range reviews {
			requestText.WriteString(fmt.Sprintf("%s - Score: %.1f\n", names[i], review.Score))
		}
	} else {
		movies, averages, err := store.GetMovies()
		if err != nil {
			status.ErrInternal(err)
			return
		}
		for i, movie := range movies {
			requestText.WriteString(fmt.Sprintf("%s - Average Score: %.1f\n", movie, averages[i]))
		}
	}

	// Send the request to ChatGPT
	recommendations, err := utils.ChatGPTRequest(requestText.String())
	if err != nil {
		status.ErrInternal(err)
		return
	}

	// Respond with the AI recommendations
	status.StatusOK(recommendations).Render(w)

}
