package main

// TMDBResponse represents the structure of the response from the TMDB API
type TMDBResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalResults int     `json:"total_results"`
	TotalPages   int     `json:"total_pages"`
}

// Movie represents the structure of a single movie in the results
type Movie struct {
	Title       string  `json:"title"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
	Adult       bool    `json:"adult"`
	ReleaseDate string  `json:"release_date"`
}

type Review struct {
	AuthorID string  `json:"author"`
	Score    float64 `json:"score"`
	Comment  string  `json:"comment"`
}
