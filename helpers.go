package main

import (
	"context"
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
	"github.com/google/generative-ai-go/genai"
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

func GeminiRequestMovieExaminationFast(movieName string) (string, error) {
	model := geminiClient.GenerativeModel("gemini-1.5-flash")
	cs := model.StartChat()

	resp, err := cs.SendMessage(context.Background(), genai.Text(movieName+" filmi hakkında ne düşünüyorsun?"))
	if err != nil {
		return "", err
	}

	result := ""

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if xo, err := json.Marshal(&part); err != nil {
					continue
				} else {
					result = string(xo)
				}
				break
			}
		}
	}

	return strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(result, "\""), "\""), "\\n", "\n"), nil
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

		timer = time.AfterFunc(300*time.Millisecond, func() {
			mu.Lock()
			defer mu.Unlock()
			f()
		})
	}
}
