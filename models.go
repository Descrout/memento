package main

type MovieAverageScore struct {
	MovieName    string  `json:"movieName"`
	AverageScore float64 `json:"averageScore"`
}

type Review struct {
	AuthorID  string  `json:"author"`
	MovieName string  `json:"movieName"`
	Score     float64 `json:"score"`
	Comment   string  `json:"comment"`
}

type MovieReviewResponse struct {
	MovieName    string   `json:"movieName"`
	AverageScore float64  `json:"averageScore"`
	Reviews      []Review `json:"reviews"`
}

// In-memory data store
var reviews = []Review{
	{AuthorID: "1", MovieName: "Inception", Score: 9.0, Comment: "Great movie!"},
	{AuthorID: "1", MovieName: "Interstellar", Score: 8.5, Comment: "Amazing visuals."},
	{AuthorID: "2", MovieName: "The Matrix", Score: 9.5, Comment: "A sci-fi classic."},
	{AuthorID: "2", MovieName: "Interstellar", Score: 7, Comment: "Not bad."},
}
