package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var store *Store

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("env could not be loaded: ", err)
	}

	discord, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("discord session could not be initialized: ", err)
	}

	store, err = NewStore()
	if err != nil {
		log.Fatal("db could not be initialized: ", err)
	}

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	discord.AddHandler(OnMessageCreated)

	err = discord.Open()
	if err != nil {
		log.Fatal("discord session could not be initialized: ", err)
	}

	// Cleanup
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, syscall.SIGQUIT)
	<-sigch

	discord.Close()
}

func OnMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	//s.ChannelMessageSend(m.ChannelID, "BURAYA BOT MESAJI")

	//addmovie Command
	if strings.HasPrefix(m.Content, "/addmovie ") {
		movieName := strings.TrimPrefix(m.Content, "/addmovie ")
		movieName = strings.TrimSpace(movieName)
		movieName = strings.ToLower(movieName) // Convert to lowercase

		alreadyExists, err := store.AddMovie(movieName)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to add movie: "+err.Error())
			return
		}

		if alreadyExists {
			s.ChannelMessageSend(m.ChannelID, "Movie already exists: "+movieName)
		} else {
			s.ChannelMessageSend(m.ChannelID, "Movie added successfully: "+movieName)
		}
	}

	//score Command
	if strings.HasPrefix(m.Content, "/score ") {
		content := strings.TrimSpace(m.Content[len("/score "):])
		firstSpace := strings.Index(content, " ")
		if firstSpace == -1 {
			s.ChannelMessageSend(m.ChannelID, "Usage: /score [score] [movie name]")
			return
		}

		scoreStr, movieName := content[:firstSpace], strings.TrimSpace(content[firstSpace+1:])
		movieName = strings.ToLower(movieName) // Convert to lowercase
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid score format. Please enter a valid number.")
			return
		}
		if score < 0 || score > 10 {
			s.ChannelMessageSend(m.ChannelID, "Score must be between 0 and 10.")
			return
		}
		if !validateScore(scoreStr) {
			s.ChannelMessageSend(m.ChannelID, "Score can only have up to two decimal places.")
			return
		}

		userID := m.Author.ID
		err = store.AddScore(movieName, score, userID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to add score: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Score of %.2f added successfully for '%s'", score, movieName))
	}

	//movie Command
	if strings.HasPrefix(m.Content, "/movie ") {
		movieName := strings.TrimSpace(m.Content[len("/movie "):])
		movieName = strings.ToLower(movieName) // Convert to lowercase
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

	//cleardb Command
	if strings.HasPrefix(m.Content, "/cleardb") {
		// Ensure the command issuer has admin privileges
		if err := store.ClearAllData(); err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to clear database: "+err.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, "Database cleared successfully.")
		}
	}

	//movies Command
	if strings.HasPrefix(m.Content, "/movies") {
		movies, err := store.ListMovies()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to retrieve movie list: "+err.Error())
			return
		}

		if len(movies) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No movies found in the database.")
			return
		}

		response := "List of movies and their average scores:\n"
		for _, movie := range movies {
			_, average, err := store.GetScores(movie) // Reuse GetScores to fetch average scores
			if err != nil {
				response += fmt.Sprintf("%s: Error retrieving scores\n", movie)
				continue
			}
			if average == 0 {
				response += fmt.Sprintf("%s: No score available\n", movie)
			} else {
				response += fmt.Sprintf("%s: %.2f\n", movie, average)
			}
		}

		s.ChannelMessageSend(m.ChannelID, response)
	}
}

// Helper function to validate decimal places
func validateScore(scoreStr string) bool {
	parts := strings.Split(scoreStr, ".")
	if len(parts) == 2 && len(parts[1]) > 2 {
		return false
	}
	return true
}

//m.Author.Username
//m.Author.ID
/*else if strings.HasPrefix(m.Content, "/comment ") {
 */

/*
/addmovie Fight Club
/movie Fight Club // Filmin scorelarını ve yorumlarını çeksin, ortalama puanı gözüksün
/comment Fight Club // Bir sonraki mesajınız bu filme yorum olarak eklenecek.
/score 8.7 Fight Club
*/

/*
Fight Club 7.9
---------------
descrout 7.6
nur 8.2
*/
