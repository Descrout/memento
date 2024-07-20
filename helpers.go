package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func InteractionAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func SanitiseCommand(content string, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(content, prefix))
}

func ValidateScore(scoreStr string) (float64, error) {
	parts := strings.Split(scoreStr, ".")
	if len(parts) == 2 && len(parts[1]) > 1 {
		return 0, fmt.Errorf("Score must be integer(eg 3) or 1 scale decimal(eg 7.5)")
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid score format: %s", err.Error())
	}

	if score < 0 || score > 10 {
		return 0, fmt.Errorf("Score must be between 0 and 10.")
	}

	return score, nil
}

func SearchMovies(query string) ([]string, error) {
	apiKey := "6ca2749b05fab4e15300c20ebf1cf782"
	baseURL := "https://api.themoviedb.org/3/search/movie"
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("query", query)

	// Create the URL with the query parameters
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make the HTTP GET request
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse the JSON response
	var tmdbResponse TMDBResponse
	err = json.Unmarshal(body, &tmdbResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	// Extract movie titles, scores, and release dates, limited to 5
	var titlesWithDetails []string
	for _, movie := range tmdbResponse.Results {
		if movie.VoteCount >= 100 && !movie.Adult {
			releaseDate := strings.Split(movie.ReleaseDate, "-")[0]
			titlesWithDetails = append(titlesWithDetails, fmt.Sprintf("%s (%s) ", movie.Title, releaseDate))
			if len(titlesWithDetails) >= 8 {
				break
			}
		}
	}

	return titlesWithDetails, nil
}
