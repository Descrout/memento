package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandFunc = func(s *discordgo.Session, m *discordgo.MessageCreate, content string)

func AddMovie(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	movieName := SanitiseMovieName(content)

	if err := store.AddMovie(movieName); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to add movie: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Movie added successfully: "+movieName)
}

func ScoreMovie(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	firstSpace := strings.Index(content, " ")
	if firstSpace == -1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: /score [score] [movie name]")
		return
	}

	scoreStr, movieName := content[:firstSpace], strings.TrimSpace(content[firstSpace+1:])

	score, err := ValidateScore(scoreStr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	movieName = SanitiseMovieName(movieName)

	if err = store.AddScore(movieName, score, m.Author.ID); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to add score: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Score of %.1f added successfully for '%s'", score, movieName))
}

func FetchMovie(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	movieName := SanitiseMovieName(content)

	scores, average, err := store.GetScores(movieName)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving scores: "+err.Error())
		return
	}

	if len(scores) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No scores available for '"+movieName+"'")
		return
	}

	// Build the response message
	response := fmt.Sprintf("%s %.1f\n---------------\n", movieName, average)
	for userID, score := range scores {
		user, err := s.User(userID) // Fetch user details
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to retrieve user details for ID: "+userID)
			continue
		}
		username := user.Username // Get the username
		response += fmt.Sprintf("%s %.1f\n", username, score)
	}

	s.ChannelMessageSend(m.ChannelID, response)
}

func ClearDB(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	// Ensure the command issuer has admin privileges
	if err := store.ClearAllData(); err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to clear database: "+err.Error())
	} else {
		s.ChannelMessageSend(m.ChannelID, "Database cleared successfully.")
	}
}
