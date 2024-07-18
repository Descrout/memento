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

var store *Store

// deneme
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

	if strings.HasPrefix(m.Content, "/addmovie ") {
		movieName := strings.TrimPrefix(m.Content, "/addmovie ")
		movieName = strings.TrimSpace(movieName)
		log.Println(movieName)
		//m.Author.Username
		//m.Author.ID
	} else if strings.HasPrefix(m.Content, "/movie ") {
	} else if strings.HasPrefix(m.Content, "/comment ") {
	} else if strings.HasPrefix(m.Content, "/score ") {
	}
}

/*
/addmovie Fight Club
/movie Fight Club // Filmin scorelarını ve yorumlarını çeksin, ortalama puanı gözüksün
/comment Fight Club // Bir sonraki mesajınız bu filme yorum olarak eklenecek.
/score 8.7 Fight Club
*/

/*
Fight Club 8.3
---------------
descrout 7.6
nur 8.2
*/
