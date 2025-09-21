package reply

import (
	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/utils/json/option"
)

type FollowUpBuilder struct {
	manager *ResponseManager
	data    api.InteractionResponseData
}

func (fb *FollowUpBuilder) Content(content string) *FollowUpBuilder {
	fb.data.Content = option.NewNullableString(content)
	return fb
}

func (fb *FollowUpBuilder) Embed(embed discord.Embed) *FollowUpBuilder {
	fb.data.Embeds = &[]discord.Embed{embed}
	return fb
}

func (fb *FollowUpBuilder) Embeds(embeds ...discord.Embed) *FollowUpBuilder {
	fb.data.Embeds = &embeds
	return fb
}

func (fb *FollowUpBuilder) Components(components discord.ContainerComponents) *FollowUpBuilder {
	fb.data.Components = &components
	return fb
}

func (fb *FollowUpBuilder) Flags(flags ...discord.MessageFlags) *FollowUpBuilder {
	for _, flag := range flags {
		fb.data.Flags |= flag
	}
	return fb
}

func (fb *FollowUpBuilder) Send() (*discord.Message, error) {
	return fb.manager.state.FollowUpInteraction(fb.manager.appID, fb.manager.token, fb.data)
}
