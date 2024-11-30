package main

import (
	"context"
	"log"
	"memento/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var (
	store      *Store
	minVal     = float64(1)
	minValZero = float64(0)
	commands   = []*discordgo.ApplicationCommand{
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
				{
					Name:         "top",
					Description:  "Limit the result count.",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minVal,
				},
			},
		},
		{
			Name:        "myreviews",
			Description: "Get the list of your reviews.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "top",
					Description:  "Limit the result count.",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minVal,
				},
			},
		},
		{
			Name:        "getreviews",
			Description: "Get the list of a specific user reviews.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "user",
					Description:  "Select a user",
					Type:         discordgo.ApplicationCommandOptionUser,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "top",
					Description:  "Limit the result count.",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minVal,
				},
			},
		},
		{
			Name:        "allmovies",
			Description: "Get the all the movies you have watched.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "top",
					Description:  "Limit the result count.",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minVal,
				},
			},
		},
		{
			Name:        "recommend",
			Description: "Get 3 movie recommendations based on your watched movies and given scores.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "personal",
					Description: "Use personal movie list and scores.",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
			},
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
				{
					Name:        "personal",
					Description: "Use personal movie list and scores.",
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Required:    true,
				},
			},
		},
		{
			Name:        "disconnect",
			Description: "After some time, disconnects you from a voice channel in 'this' server.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "hours",
					Description:  "Hours",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minValZero,
				},
				{
					Name:         "minutes",
					Description:  "Minutes",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minValZero,
				},
				{
					Name:         "seconds",
					Description:  "Seconds",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minValZero,
				},
			},
		},
		{
			Name:        "cancel",
			Description: "Cancels your disconnect request.",
			Type:        discordgo.ChatApplicationCommand,
		},
		{
			Name:        "similarusers",
			Description: "Get the list of similar users for a specific user.",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "user",
					Description:  "Select a user",
					Type:         discordgo.ApplicationCommandOptionUser,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "top",
					Description:  "Limit the result count.",
					Type:         discordgo.ApplicationCommandOptionInteger,
					Required:     false,
					Autocomplete: false,
					MinValue:     &minVal,
				},
			},
		},
	}

	commandFuncs = map[string]CommandFunc{
		"review":       ReviewCommand,
		"movie":        MovieCommand,
		"allmovies":    GetMoviesCommand,
		"delete":       DeleteCommand,
		"examine":      ExamineCommand,
		"myreviews":    MyReviewsCommand,
		"getreviews":   GetReviewsCommand,
		"recommend":    RecommendCommand,
		"disconnect":   DisconnectCommand,
		"cancel":       CancelCommand,
		"similarusers": SimilarUsersCommand,
	}

	debouncers = NewMutexMap[string, utils.DebounceFunc]()

	disconnectTimers = NewMutexMap[string, *time.Timer]()

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

	guildID := os.Getenv("GUILD_ID")
	createdCommands, err := discord.ApplicationCommandBulkOverwrite(discord.State.User.ID, guildID, commands)
	if err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	// Webserver
	// router := chi.NewRouter()
	// router.Use(middleware.RequestID)
	// router.Use(middleware.RealIP)
	// router.Use(middleware.Logger)
	// router.Use(middleware.Recoverer)
	// router.Use(middleware.Timeout(60 * time.Second))

	// router.Get("/myreviews", GetReviewsByAuthorID)
	// router.Get("/allmovies", GetAllMovies)
	// router.Get("/movie", GetReviewsByMovieName)

	// router.Get("/allmovies", GetAllMovies)

	// server := http.Server{
	// 	Addr:    ":8080",
	// 	Handler: router,
	// }
	// defer server.Close()
	// go server.ListenAndServe()

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
