package reply

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/pkg/constants/emojis"
)

type ResponseBuilder struct {
	e          *events.ApplicationCommandInteractionCreate
	content    *string
	embeds     []discord.Embed
	components []discord.LayoutComponent
	flags      discord.MessageFlags
	ephemeral  bool
	files      []*discord.File
}

func New(e *events.ApplicationCommandInteractionCreate) *ResponseBuilder {
	return &ResponseBuilder{
		e: e,
	}
}

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

func (r *ResponseBuilder) File(file *discord.File) *ResponseBuilder {
	r.files = append(r.files, file)
	return r
}

func (r *ResponseBuilder) Embed(embed discord.Embed) *ResponseBuilder {
	r.embeds = append(r.embeds, embed)
	return r
}

func (r *ResponseBuilder) Component(component discord.LayoutComponent) *ResponseBuilder {
	r.components = append(r.components, component)
	return r
}

func (r *ResponseBuilder) Ephemeral() *ResponseBuilder {
	r.ephemeral = true
	return r
}

func (r *ResponseBuilder) Defer() error {
	return r.e.DeferCreateMessage(r.ephemeral)
}

func (r *ResponseBuilder) FollowUp() error {
	msg := discord.MessageCreate{
		Content:         *r.content,
		Embeds:          r.embeds,
		Files:           r.files,
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

func (r *ResponseBuilder) Send() error {
	_, err := r.e.Client().Rest.UpdateInteractionResponse(
		r.e.ApplicationID(),
		r.e.Token(),
		discord.MessageUpdate{
			Components:      &r.components,
			Content:         r.content,
			Embeds:          &r.embeds,
			Files:           r.files,
			Flags:           &r.flags,
			AllowedMentions: &discord.AllowedMentions{},
		},
	)
	return err
}

func (r *ResponseBuilder) Edit() error {
	_, err := r.e.Client().Rest.UpdateInteractionResponse(
		r.e.ApplicationID(),
		r.e.Token(),
		discord.MessageUpdate{
			Components:      &r.components,
			Flags:           &r.flags,
			Content:         r.content,
			Files:           r.files,
			Embeds:          &r.embeds,
			AllowedMentions: &discord.AllowedMentions{},
		},
	)
	return err
}

func QuickEmbed(title, description string) discord.Embed {
	return discord.NewEmbedBuilder().
		SetTitle(title).
		SetDescription(description).
		SetColor(0x00ADD8).
		Build()
}

func Error(e *events.ApplicationCommandInteractionCreate, err error) error {
	embed := QuickEmbed(fmt.Sprintf("%s error", emojis.EmojiCross), err.Error())
	embed.Color = 0xE74C3C

	return New(e).
		Embed(embed).
		Send()
}
