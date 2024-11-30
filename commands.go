package main

import (
	"fmt"
	"log"
	"memento/models"
	"memento/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type CommandFunc = func(s *discordgo.Session, i *discordgo.InteractionCreate)

func RespondToInteraction(s *discordgo.Session, interaction *discordgo.Interaction, message string) {
	s.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func ReviewCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		movieName := strings.TrimSpace(data.Options[0].StringValue())
		score := data.Options[1].FloatValue()
		comment := data.Options[2].StringValue()

		author := utils.InteractionAuthor(i.Interaction)

		if err := store.AddReview(movieName, &models.Review{
			AuthorID: author.ID,
			Score:    score,
			Comment:  comment,
		}); err != nil {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("Review could not be added: %s", err.Error()))
			return
		}

		RespondToInteraction(s, i.Interaction, fmt.Sprintf("**%s** reviewed ``%s`` ``%.2f``\n```%s```", author.Username, movieName, score, comment))
		break
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := utils.InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, utils.Debouncer())
		debounce(func() {
			names := []string{}
			namesReviewed, err := store.SearchMovies(name)
			if err == nil {
				names = append(names, namesReviewed...)
			}

			diff := 8 - len(names)
			if diff > 0 {
				namesTmdb, err := utils.SearchMovies(name, diff)
				if err == nil {
					names = append(names, namesTmdb...)
				}
			}

			names = utils.FilterUnique(names)

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
		top := utils.GetTop(data.Options)

		movieName := strings.TrimSpace(data.Options[0].StringValue())

		reviews, avg, err := store.GetReviews(movieName)
		if err != nil {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("Reviews could not be fetched: %s", err.Error()))
			return
		}

		batcher := utils.NewBatcher(2000)
		batcher.Add(fmt.Sprintf("# %s ``[%.2f]``\n", movieName, avg))
		for j, review := range reviews {
			user, err := s.User(review.AuthorID)
			if err != nil {
				continue
			}
			batcher.Add(fmt.Sprintf("%d. **%s** **``(%.2f)``** - ``\"%s\"``\n", j+1, user.Username, review.Score, review.Comment))
			if top != 0 && j+1 >= int(top) {
				break
			}
		}

		batcher.Send(s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := utils.InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, utils.Debouncer())
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
		data := i.ApplicationCommandData()
		top := utils.GetTop(data.Options)
		movies, averages, err := store.GetMovies()
		if err != nil {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("Movies could not be fetched: %s", err.Error()))
			return
		}

		batcher := utils.NewBatcher(2000)

		for j, movieName := range movies {
			batcher.Add(fmt.Sprintf("%d. **%s** ``%.2f``\n", j+1, movieName, averages[j]))

			if top != 0 && j+1 >= int(top) {
				break
			}
		}

		batcher.Send(s, i)
	}
}

func DeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		movieName := strings.TrimSpace(data.Options[0].StringValue())
		author := utils.InteractionAuthor(i.Interaction)

		err := store.DeleteReview(movieName, author.ID)
		if err != nil {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("Review could not be deleted: %s", err.Error()))
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

		RespondToInteraction(s, i.Interaction, result)
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		name := strings.TrimSpace(data.Options[0].StringValue())

		author := utils.InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, utils.Debouncer())
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
		personal := data.Options[1].BoolValue()
		var requestText strings.Builder

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		requestText.WriteString("Sana film listesi ve onlara verdiğim puanları vereceğim. Bu puanlardan yola çıkarak sence " + movieName + " filmi hakkında ne düşünürüm? 2000 karakterden az cevap ver lütfen.\n\nListe:\n")

		if personal {
			reviews, names, err := store.GetReviewsByUser(utils.InteractionAuthor(i.Interaction).ID)
			if err != nil {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: fmt.Sprintf("AI Examination failed: %s", err.Error()),
				})
				return
			}
			for i, review := range reviews {
				requestText.WriteString(fmt.Sprintf("%s - Score: %.2f\n", names[i], review.Score))
			}
		} else {
			movies, averages, err := store.GetMovies()
			if err != nil {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: fmt.Sprintf("AI Examination failed: %s", err.Error()),
				})
				return
			}
			for i, movie := range movies {
				requestText.WriteString(fmt.Sprintf("%s - Average Score: %.2f\n", movie, averages[i]))
			}
		}

		examination, err := utils.ChatGPTRequest(requestText.String())

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

		author := utils.InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, utils.Debouncer())
		debounce(func() {
			names := []string{}
			namesReviewed, err := store.SearchMovies(name)
			if err == nil {
				names = append(names, namesReviewed...)
			}

			diff := 8 - len(names)
			if diff > 0 {
				namesTmdb, err := utils.SearchMovies(name, diff)
				if err == nil {
					names = append(names, namesTmdb...)
				}
			}

			names = utils.FilterUnique(names)

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

func MyReviewsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()
	top := utils.GetTop(data.Options)
	author := utils.InteractionAuthor(i.Interaction)
	reviews, names, err := store.GetReviewsByUser(author.ID)
	if err != nil {
		RespondToInteraction(s, i.Interaction, fmt.Sprintf("Failed to fetch reviews: %s", err.Error()))
		return
	}

	if len(reviews) == 0 {
		RespondToInteraction(s, i.Interaction, "You haven't reviewed any movies yet.")
		return
	}

	batcher := utils.NewBatcher(2000)
	batcher.Add(fmt.Sprintf("# %s ``[%.2f]``\n", author.Username, utils.AverageScore(reviews)))

	for j, review := range reviews {
		batcher.Add(fmt.Sprintf("%d. **%s** **``(%.2f)``** - ``\"%s\"``\n", j+1, names[j], review.Score, review.Comment))
		if top != 0 && j+1 >= int(top) {
			break
		}
	}

	batcher.Send(s, i)
}

func RecommendCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		personal := data.Options[0].BoolValue()
		author := utils.InteractionAuthor(i.Interaction)

		var requestText strings.Builder
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		requestText.WriteString("Sana film listesi ve onlara verdiğim puanları vereceğim. Bunlara göre bana beğenebileceğim 3 film öner. 2000 karakterden az cevap ver lütfen. Liste:\n\n")

		if personal {
			reviews, names, err := store.GetReviewsByUser(author.ID)
			if err != nil || len(reviews) == 0 {
				responseMessage := "You haven't reviewed any movies yet, so recommendations cannot be provided."
				if err != nil {
					responseMessage = fmt.Sprintf("AI Recommendation failed: %s", err.Error())
				}
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: responseMessage,
				})
				return
			}

			for i, review := range reviews {
				requestText.WriteString(fmt.Sprintf("%s - Score: %.2f\n", names[i], review.Score))
			}
		} else {
			movies, averages, err := store.GetMovies()
			if err != nil {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: fmt.Sprintf("AI Recommendation failed: %s", err.Error()),
				})
				return
			}
			for i, movie := range movies {
				requestText.WriteString(fmt.Sprintf("%s - Average Score: %.2f\n", movie, averages[i]))
			}
		}

		recommendations, err := utils.ChatGPTRequest(requestText.String())
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("AI Recommendation failed: %s", err.Error()),
			})
		} else {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: recommendations,
			})
		}
	}
}

func GetReviewsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		top := utils.GetTop(data.Options)
		author := data.Options[0].UserValue(s)
		reviews, names, err := store.GetReviewsByUser(author.ID)
		if err != nil {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("Failed to fetch reviews: %s", err.Error()))
			return
		}

		if len(reviews) == 0 {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("%s haven't reviewed any movies yet.", author.Username))
			return
		}

		batcher := utils.NewBatcher(2000)
		batcher.Add(fmt.Sprintf("# %s ``[%.2f]``\n", author.Username, utils.AverageScore(reviews)))

		for j, review := range reviews {
			batcher.Add(fmt.Sprintf("%d. **%s** **``(%.2f)``** - ``\"%s\"``\n", j+1, names[j], review.Score, review.Comment))

			if top != 0 && j+1 >= int(top) {
				break
			}
		}

		batcher.Send(s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		userInput := strings.ToLower(strings.TrimSpace(data.Options[0].StringValue()))

		author := utils.InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, utils.Debouncer())
		debounce(func() {
			members, err := s.GuildMembers(i.GuildID, "", 100)
			if err != nil {
				log.Println("Error fetching guild members:", err)
				return
			}

			suggestions := []*discordgo.ApplicationCommandOptionChoice{}
			for _, member := range members {
				user := member.User
				nameMatches := strings.Contains(strings.ToLower(user.Username), userInput)
				nickMatches := strings.Contains(strings.ToLower(member.Nick), userInput)

				if nameMatches || nickMatches {
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  fmt.Sprintf("%s#%s", user.Username, user.Discriminator),
						Value: user.ID,
					})
				}

				if len(suggestions) >= 25 {
					break
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: suggestions,
				},
			})
		})
	}
}

func DisconnectCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	duration := time.Duration(0)
	for _, opt := range data.Options {
		duration += time.Duration(opt.IntValue()) * utils.GetDurationFromStr(opt.Name)
	}

	if duration == 0 {
		RespondToInteraction(s, i.Interaction, "Duration cannot be zero.")
		return
	}

	author := utils.InteractionAuthor(i.Interaction)
	guildID := i.Interaction.GuildID
	guildName := "this"
	g, err := s.Guild(guildID)
	if err == nil {
		guildName = g.Name
	}

	key := fmt.Sprintf("%s|%s", author.ID, guildID)

	disconnectTimers.Lock()

	timerBefore := disconnectTimers.Get(key)
	if timerBefore != nil {
		timerBefore.Stop()
	}

	disconnectTimers.SetNonBlocking(key, time.AfterFunc(duration, func() {
		s.GuildMemberMove(guildID, author.ID, nil)
		disconnectTimers.Delete(key)
	}))

	disconnectTimers.Unlock()

	RespondToInteraction(s, i.Interaction, fmt.Sprintf("I will disconnect you from a voice channel in **%s** server, after ``%s``", guildName, duration.String()))
}

func CancelCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	author := utils.InteractionAuthor(i.Interaction)
	guildID := i.Interaction.GuildID
	guildName := "this"
	g, err := s.Guild(guildID)
	if err == nil {
		guildName = g.Name
	}

	key := fmt.Sprintf("%s|%s", author.ID, guildID)

	timerBefore := disconnectTimers.Get(key)
	if timerBefore != nil {
		timerBefore.Stop()
		disconnectTimers.Delete(key)
		RespondToInteraction(s, i.Interaction, fmt.Sprintf("Disconnect timer successfuly cancelled for **%s** server.", guildName))
	} else {
		RespondToInteraction(s, i.Interaction, fmt.Sprintf("You did not set any disconnect timers on **%s** server.", guildName))
	}

}

func SimilarUsersCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		top := utils.GetTop(data.Options)
		author := data.Options[0].UserValue(s)
		similars, err := store.GetUserSimilarities(author.ID)
		if err != nil {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("Failed to fetch similarities: %s", err.Error()))
			return
		}

		if len(similars) == 0 {
			RespondToInteraction(s, i.Interaction, fmt.Sprintf("We couldn't find any similar users to %s", author.Username))
			return
		}

		batcher := utils.NewBatcher(2000)
		batcher.Add(fmt.Sprintf("# Similar users to ``%s``\n", author.Username))

		for j, similar := range similars {
			user, err := s.User(similar.UserID)
			if err == nil {
				batcher.Add(fmt.Sprintf("%d. **%s** **``(%f)``**\n", j+1, user.Username, similar.Similarity))
			}
			if top != 0 && j+1 >= int(top) {
				break
			}
		}

		batcher.Send(s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		if !data.Options[0].Focused {
			return
		}

		userInput := strings.ToLower(strings.TrimSpace(data.Options[0].StringValue()))

		author := utils.InteractionAuthor(i.Interaction)
		debounce := debouncers.SetIfNotExists(author.ID, utils.Debouncer())
		debounce(func() {
			members, err := s.GuildMembers(i.GuildID, "", 100)
			if err != nil {
				log.Println("Error fetching guild members:", err)
				return
			}

			suggestions := []*discordgo.ApplicationCommandOptionChoice{}
			for _, member := range members {
				user := member.User
				nameMatches := strings.Contains(strings.ToLower(user.Username), userInput)
				nickMatches := strings.Contains(strings.ToLower(member.Nick), userInput)

				if nameMatches || nickMatches {
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  fmt.Sprintf("%s#%s", user.Username, user.Discriminator),
						Value: user.ID,
					})
				}

				if len(suggestions) >= 25 {
					break
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: suggestions,
				},
			})
		})
	}
}
