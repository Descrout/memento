package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

func InteractionAuthor(i *discordgo.Interaction) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func SearchMovies(query string, searchCount int) ([]string, error) {
	baseURL := "https://api.themoviedb.org/3/search/movie"
	params := url.Values{}
	params.Add("api_key", os.Getenv("TMDB_API_KEY"))
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
			if len(titlesWithDetails) >= searchCount {
				break
			}
		}
	}

	return titlesWithDetails, nil
}

type DebounceFunc = func(f func())

func Debouncer() DebounceFunc {
	var mu sync.Mutex
	var timer *time.Timer

	return func(f func()) {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(500*time.Millisecond, func() {
			mu.Lock()
			defer mu.Unlock()
			f()
		})
	}
}
