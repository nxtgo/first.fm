package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// ResponseBuilder lets you fluently build a reply
type ResponseBuilder struct {
	e         *events.ApplicationCommandInteractionCreate
	content   *string
	embeds    []discord.Embed
	ephemeral bool
}

// Reply starts a new ResponseBuilder for a deferred interaction
func Reply(e *events.ApplicationCommandInteractionCreate) *ResponseBuilder {
	return &ResponseBuilder{
		e: e,
	}
}

// Content sets the message content
func (r *ResponseBuilder) Content(msg string, a ...any) *ResponseBuilder {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	r.content = &msg
	return r
}

// Embed adds an embed
func (r *ResponseBuilder) Embed(embed discord.Embed) *ResponseBuilder {
	r.embeds = append(r.embeds, embed)
	return r
}

// Ephemeral sets the reply to be ephemeral (visible only to the user)
func (r *ResponseBuilder) Ephemeral() *ResponseBuilder {
	r.ephemeral = true
	return r
}

// Defer defers the interaction reply (must be called within 3s)
func (r *ResponseBuilder) Defer() error {
	return r.e.DeferCreateMessage(r.ephemeral)
}

// Send edits the interaction response (after Defer was called)
func (r *ResponseBuilder) Send() error {
	_, err := r.e.Client().Rest().UpdateInteractionResponse(
		r.e.ApplicationID(),
		r.e.Token(),
		discord.MessageUpdate{
			Content:         r.content,
			Embeds:          &r.embeds,
			AllowedMentions: &discord.AllowedMentions{},
		},
	)
	return err
}

// Edit edits the original deferred response
func (r *ResponseBuilder) Edit() error {
	builder := discord.NewMessageCreateBuilder()
	if r.content != nil {
		builder.SetContent(*r.content)
	}
	if len(r.embeds) > 0 {
		builder.AddEmbeds(r.embeds...)
	}

	_, err := r.e.Client().Rest().UpdateInteractionResponse(
		r.e.ApplicationID(),
		r.e.Token(),
		discord.MessageUpdate{
			Content:         r.content,
			Embeds:          &r.embeds,
			AllowedMentions: &discord.AllowedMentions{},
		},
	)
	return err
}

// QuickEmbed helper
func QuickEmbed(title, description string, color int) discord.Embed {
	return discord.NewEmbedBuilder().
		SetTitle(title).
		SetDescription(description).
		SetColor(color).
		Build()
}

// ErrorReply sends an ephemeral error embed with a red color
func ErrorReply(e *events.ApplicationCommandInteractionCreate, message string) error {
	embed := QuickEmbed("❌ error", message, 0xE74C3C)
	return Reply(e).
		Embed(embed).
		Send()
}
