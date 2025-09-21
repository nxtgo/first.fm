package reply

import (
	"context"
	"fmt"
	"time"

	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/state"
	"go.fm/pkg/emojis"
	"go.fm/zlog"
)

type ResponseManager struct {
	state       *state.State
	interaction *discord.InteractionEvent
	token       string
	appID       discord.AppID
	deferred    bool
	responded   bool
}

func New(s *state.State, i *discord.InteractionEvent) *ResponseManager {
	return &ResponseManager{
		state:       s,
		interaction: i,
		token:       i.Token,
		appID:       i.AppID,
	}
}

func (rm *ResponseManager) Quick(content string, flags ...discord.MessageFlags) error {
	builder := rm.Reply().Content(content)
	if len(flags) > 0 {
		builder = builder.Flags(flags...)
	}
	return builder.Send()
}

func (rm *ResponseManager) QuickEmbed(embed discord.Embed, flags ...discord.MessageFlags) error {
	builder := rm.Reply().Embed(embed)
	if len(flags) > 0 {
		builder = builder.Flags(flags...)
	}
	return builder.Send()
}

func (rm *ResponseManager) AutoDefer(fn func(edit *EditBuilder) error, flags ...discord.MessageFlags) error {
	deferred := rm.Defer(flags...)
	if deferred.Error() != nil {
		return deferred.Error()
	}

	editBuilder := deferred.Edit().Flags(flags...)
	err := fn(editBuilder)

	if err != nil {
		zlog.Log.Error(err.Error())
		_, err := editBuilder.Clear().Embed(ErrorEmbed(err.Error())).Send()
		return err
	}

	return nil
}

func ErrorEmbed(description string) discord.Embed {
	return discord.Embed{
		Description: fmt.Sprintf("%s %s", emojis.EmojiCross, description),
		Color:       0xFF0000,
	}
}

func SuccessEmbed(description string) discord.Embed {
	return discord.Embed{
		Description: fmt.Sprintf("%s %s", emojis.EmojiCheck, description),
		Color:       0x00FF00,
	}
}

func InfoEmbed(description string) discord.Embed {
	return discord.Embed{
		Description: fmt.Sprintf("%s %s", emojis.EmojiChat, description),
		Color:       0x0099FF,
	}
}

func WithTimeout(ctx context.Context, timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
