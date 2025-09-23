package reply

import (
	"context"
	"fmt"
	"time"

	"github.com/nxtgo/arikawa/v3/discord"
	"go.fm/internal/bot/discord/emojis"
)

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
