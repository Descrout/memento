package utils

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Batcher struct {
	builder   *strings.Builder
	messages  []string
	batchSize int
}

func NewBatcher(batchSize int) *Batcher {
	return &Batcher{
		builder:   &strings.Builder{},
		messages:  []string{},
		batchSize: batchSize,
	}
}

func (b *Batcher) Add(entry string) {
	if b.builder.Len()+len(entry) >= b.batchSize {
		b.messages = append(b.messages, b.builder.String())
		b.builder.Reset()
	}

	b.builder.WriteString(entry)
}

func (b *Batcher) flush() {
	if b.builder.Len() > 0 {
		b.messages = append(b.messages, b.builder.String())
		b.builder.Reset()
	}
}

func (b *Batcher) Send(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	b.flush()

	for idx, msg := range b.messages {
		var err error
		if idx == 0 {
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg,
				},
			})
		} else {
			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: msg,
			})
		}

		if err != nil {
			return err
		}
	}

	return nil
}
