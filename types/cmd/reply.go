package cmd

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// ResponseBuilder lets you fluently build a reply
type ResponseBuilder struct {
	e          *events.ApplicationCommandInteractionCreate
	content    *string
	embeds     []discord.Embed
	components []discord.LayoutComponent
	flags      discord.MessageFlags
	ephemeral  bool
}

// Reply starts a new ResponseBuilder for a deferred interaction
func (ctx *CommandContext) Reply(e *events.ApplicationCommandInteractionCreate) *ResponseBuilder {
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

func (r *ResponseBuilder) Flags(flags discord.MessageFlags) *ResponseBuilder {
	r.flags = flags
	return r
}

// Embed adds an embed
func (r *ResponseBuilder) Embed(embed discord.Embed) *ResponseBuilder {
	r.embeds = append(r.embeds, embed)
	return r
}

// Component adds a component
func (r *ResponseBuilder) Component(component discord.LayoutComponent) *ResponseBuilder {
	r.components = append(r.components, component)
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

// Send a followUp the interaction response
func (r *ResponseBuilder) FollowUp() error {
	msg := discord.MessageCreate{
		Content:         *r.content,
		Embeds:          r.embeds,
		AllowedMentions: &discord.AllowedMentions{},
	}
	if r.ephemeral {
		msg.Flags.Add(discord.MessageFlagEphemeral)
	}

	_, err := r.e.Client().Rest.CreateFollowupMessage(
		r.e.ApplicationID(),
		r.e.Token(),
		msg,
	)
	return err
}

// Send edits the interaction response (after Defer was called)
func (r *ResponseBuilder) Send() error {
	_, err := r.e.Client().Rest.UpdateInteractionResponse(
		r.e.ApplicationID(),
		r.e.Token(),
		discord.MessageUpdate{
			Components: &r.components,
			Content:    r.content,
			Embeds:     &r.embeds,
			Flags:      &r.flags,
		},
	)
	return err
}

// Edit edits the original deferred response
func (r *ResponseBuilder) Edit() error {
	_, err := r.e.Client().Rest.UpdateInteractionResponse(
		r.e.ApplicationID(),
		r.e.Token(),
		discord.MessageUpdate{
			Components: &r.components,
			Flags:      &r.flags,
			Content:    r.content,
			Embeds:     &r.embeds,
		},
	)
	return err
}

// QuickEmbed helper
func (_ *CommandContext) QuickEmbed(title, description string) discord.Embed {
	return discord.NewEmbedBuilder().
		SetTitle(title).
		SetDescription(description).
		SetColor(0x00ADD8).
		Build()
}

// ErrorReply sends an ephemeral error embed with a red color
func (c *CommandContext) Error(e *events.ApplicationCommandInteractionCreate, message string) error {
	embed := c.QuickEmbed("❌ error", message)
	embed.Color = 0xE74C3C

	return c.Reply(e).
		Embed(embed).
		Send()
}
