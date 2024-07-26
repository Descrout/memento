package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/myreviews", GetReviewsByAuthorID)
	mux.HandleFunc("/allmovies", GetMoviesWithAverageScores)
	mux.HandleFunc("/movie", GetReviewsByMovieName)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	defer server.Close()
	server.ListenAndServe() //go
}

// GetReviewsByAuthorID retrieves reviews by a specific user
func GetReviewsByAuthorID(w http.ResponseWriter, r *http.Request) {
	// Get the author ID from the URL parameters
	AuthorID := r.URL.Query().Get("AuthorID")

	// Filter reviews by author ID
	var userReviews []Review
	for _, review := range reviews {
		if review.AuthorID == AuthorID {
			userReviews = append(userReviews, review)
		}
	}

	// Send the reviews as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userReviews)
}

// GetMoviesWithAverageScores retrieves all movies with their average scores
func GetMoviesWithAverageScores(w http.ResponseWriter, r *http.Request) {
	// Map to hold the total score and count of reviews for each movie
	movieScores := make(map[string]struct {
		totalScore float64
		count      int
	})

	// Calculate total scores and count reviews for each movie
	for _, review := range reviews {
		if _, exists := movieScores[review.MovieName]; !exists {
			movieScores[review.MovieName] = struct {
				totalScore float64
				count      int
			}{}
		}
		movieScores[review.MovieName] = struct {
			totalScore float64
			count      int
		}{
			totalScore: movieScores[review.MovieName].totalScore + review.Score,
			count:      movieScores[review.MovieName].count + 1,
		}
	}

	// Calculate average scores
	var movies []MovieAverageScore
	for movieName, scoreData := range movieScores {
		averageScore := scoreData.totalScore / float64(scoreData.count)
		movies = append(movies, MovieAverageScore{
			MovieName:    movieName,
			AverageScore: averageScore,
		})
	}

	// Send the movies with average scores as a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// GetReviewsByMovieName retrieves reviews and average score for a specific movie
func GetReviewsByMovieName(w http.ResponseWriter, r *http.Request) {
	// Get the movie name from the URL query parameters
	movieName := r.URL.Query().Get("movieName")

	// Filter reviews by movie name
	var movieReviews []Review
	var totalScore float64
	for _, review := range reviews {
		if review.MovieName == movieName {
			movieReviews = append(movieReviews, review)
			totalScore += review.Score
		}
	}

	// Calculate the average score
	var averageScore float64
	if len(movieReviews) > 0 {
		averageScore = totalScore / float64(len(movieReviews))
	}

	//Create a response object
	response := MovieReviewResponse{
		MovieName:    movieName,
		AverageScore: averageScore,
		Reviews:      movieReviews,
	}

	// Send the response as a JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET POST PUT PATCH DELETE
