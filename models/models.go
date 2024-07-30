package models

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

// Create a response structure
type GetReviewsByUserResponse struct {
	MovieName string           `json:"movie_name"`
	Reviews   []ReviewResponse `json:"reviews"`
}

type ReviewResponse struct {
	AuthorID string  `json:"author_id"`
	Score    float64 `json:"score"`
	Comment  string  `json:"comment"`
}

type MovieResponse struct {
	MovieName    string  `json:"movie_name"`
	AverageScore float64 `json:"average_score"`
}

type MovieWithReviewsResponse struct {
	MovieName    string           `json:"movie_name"`
	AverageScore float64          `json:"average_score"`
	Reviews      []ReviewResponse `json:"reviews"`
}

type ChatGPTError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   any    `json:"param"`
		Code    any    `json:"code"`
	} `json:"error"`
}

type ChatGPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     any    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatGPTChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

type AddReviewRequest struct {
	AuthorID  string  `json:"author_id"`
	MovieName string  `json:"movie"`
	Score     float64 `json:"score"`
	Comment   string  `json:"comment"`
}

type ExamineRequest struct {
	AuthorID  string `json:"author_id"`
	MovieName string `json:"movie_name"`
	Personal  bool   `json:"personal"`
}
