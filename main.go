package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	store    *Store
	commands = map[string]CommandFunc{
		"/addmovie ": AddMovie,
		"/score ":    ScoreMovie,
		"/movie ":    FetchMovie,
		"/cleardb":   ClearDB,
	}
)

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
	defer store.Close()

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	discord.AddHandler(OnMessageCreated)

	err = discord.Open()
	if err != nil {
		log.Fatal("discord session could not be initialized: ", err)
	}
	defer discord.Close()

	// Cleanup
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, syscall.SIGQUIT)
	<-sigch
}

func OnMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	for prefix, commandFunc := range commands {
		if strings.HasPrefix(m.Content, prefix) {
			commandFunc(s, m, SanitiseCommand(m.Content, prefix))
			break
		}
	}
}
