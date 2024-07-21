package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var (
	store    *Store
	minVal   = float64(1)
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "review",
			Description: "Set a review for a movie.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "movie",
					Description:  "Name of the movie.",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "score",
					Description:  "Score of the movie.",
					Type:         discordgo.ApplicationCommandOptionNumber,
					Required:     true,
					Autocomplete: false,
					MinValue:     &minVal,
					MaxValue:     10,
				},
				{
					Name:         "comment",
					Description:  "Small comment for the movie.",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: false,
					MaxLength:    150,
				},
			},
		},
		{
			Name:        "movie",
			Description: "Get the reviews for a movie.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "movie",
					Description:  "Name of the movie",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "allmovies",
			Description: "Get the all the movies you have watched.",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "delete",
			Description: "Delete your review about this movie.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "movie",
					Description:  "Name of the movie",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "examine",
			Description: "Get the AI review for a movie.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "movie",
					Description:  "Name of the movie",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
	}

	commandFuncs = map[string]CommandFunc{
		"review":    ReviewCommand,
		"movie":     MovieCommand,
		"allmovies": GetMoviesCommand,
		"delete":    DeleteCommand,
		"examine":   ExamineCommand,
	}

	debouncers = NewMutexMap[string, DebounceFunc]()

	geminiClient *genai.Client
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("env could not be loaded: ", err)
	}

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal("discord session could not be initialized: ", err)
	}

	geminiClient, err = genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal("gemini could not be initialized:", err)
	}
	defer geminiClient.Close()

	store, err = NewStore()
	if err != nil {
		log.Fatal("db could not be initialized: ", err)
	}
	defer store.Close()

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s", r.User.String())
	})

	//discord.AddHandler(OnMessageCreated)
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandFuncs[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = discord.Open()
	if err != nil {
		log.Fatal("discord session could not be initialized: ", err)
	}
	defer discord.Close()

	guildID := "1230981851557134396"
	createdCommands, err := discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, guildID, commands)
	if err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	// Cleanup
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt, syscall.SIGQUIT)
	<-sigch

	for _, cmd := range createdCommands {
		err := discord.ApplicationCommandDelete(discord.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Fatalf("Cannot delete %q command: %v", cmd.Name, err)
		}
	}
}

// func OnMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
// 	if m.Author.ID == s.State.User.ID {
// 		return
// 	}

// 	log.Println(m.Author.Username, ":", m.Content)
// }
