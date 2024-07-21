package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandFunc = func(s *discordgo.Session, i *discordgo.InteractionCreate)

func ReviewCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		movieName := strings.TrimSpace(data.Options[0].StringValue())
		score := data.Options[1].FloatValue()
		comment := data.Options[2].StringValue()

		author := InteractionAuthor(i.Interaction)

		if err := store.AddReview(movieName, &Review{
			AuthorID: author.ID,
			Score:    score,
			Comment:  comment,
		}); err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Review could not be added: %s", err.Error()),
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("**%s** reviewed ``%s`` ``%.1f``\n```%s```", author.Username, movieName, score, comment),
				//Content: fmt.Sprintf("A review added for the movie **``%s``** by **%s:** ``(%.1f)``  **-** ``\"%s\"`` ", movieName, i.Interaction.Member.Nick, score, comment),
			},
		})

		break
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, Debouncer())
		debounce(func() {
			names := []string{}
			namesReviewed, err := store.SearchMovies(name)
			if err == nil {
				names = append(names, namesReviewed...)
			}

			diff := 8 - len(names)
			if diff > 0 {
				namesTmdb, err := SearchMovies(name, diff)
				if err == nil {
					names = append(names, namesTmdb...)
				}
			}

			choices := []*discordgo.ApplicationCommandOptionChoice{}
			for _, name := range names {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  name,
					Value: name,
				})
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
		})

	}
}

func MovieCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		movieName := strings.TrimSpace(data.Options[0].StringValue())

		reviews, avg, err := store.GetReviews(movieName)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Reviews could not be fetched: %s", err.Error()),
				},
			})
			return
		}

		result := fmt.Sprintf("# %s ``[%.1f]``\n", movieName, avg)

		for _, review := range reviews {
			user, err := s.User(review.AuthorID)
			if err != nil {
				continue
			}
			result += fmt.Sprintf("- **%s** **``(%.1f)``** - ``\"%s\"``\n", user.Username, review.Score, review.Comment)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: result,
			},
		})
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, Debouncer())
		debounce(func() {
			choices := []*discordgo.ApplicationCommandOptionChoice{}
			names, err := store.SearchMovies(name)
			if err == nil {
				for _, name := range names {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  name,
						Value: name,
					})
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
		})
	}
}

func GetMoviesCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		movies, averages, err := store.GetMovies()
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Movies could not be fetched: %s", err.Error()),
				},
			})
			return
		}

		result := ""

		for i, movieName := range movies {
			result += fmt.Sprintf("- **%s** ``%.1f``\n", movieName, averages[i])
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: result,
			},
		})
	}
}

func DeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		movieName := strings.TrimSpace(data.Options[0].StringValue())
		author := InteractionAuthor(i.Interaction)

		err := store.DeleteReview(movieName, author.ID)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Review could not be deleted: %s", err.Error()),
				},
			})
			return
		}

		movieDeleted := false
		count, err := store.GetReviewCount(movieName)
		if err == nil && count == 0 {
			err = store.DeleteMovie(movieName)
			if err == nil {
				movieDeleted = true
			}
		}

		result := "Review deleted successfuly."
		if movieDeleted {
			result += "\nMovie has been deleted because no review left."
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: result,
			},
		})
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, Debouncer())
		debounce(func() {
			choices := []*discordgo.ApplicationCommandOptionChoice{}
			names, err := store.SearchMovies(name)
			if err == nil {
				for _, name := range names {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  name,
						Value: name,
					})
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
		})
	}
}

func ExamineCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		movieName := strings.TrimSpace(data.Options[0].StringValue())

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		examination, err := GeminiRequestMovieExaminationFast(movieName)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("AI Examination failed: %s", err.Error()),
			})
		} else {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: examination,
			})
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, Debouncer())
		debounce(func() {
			names := []string{}
			namesReviewed, err := store.SearchMovies(name)
			if err == nil {
				names = append(names, namesReviewed...)
			}

			diff := 8 - len(names)
			if diff > 0 {
				namesTmdb, err := SearchMovies(name, diff)
				if err == nil {
					names = append(names, namesTmdb...)
				}
			}

			choices := []*discordgo.ApplicationCommandOptionChoice{}
			for _, name := range names {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  name,
					Value: name,
				})
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
		})
	}
}
