package reply

import (
	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/utils/json"
	"github.com/nxtgo/arikawa/v3/utils/json/option"
)

type EditBuilder struct {
	manager *ResponseManager
	data    api.EditInteractionResponseData
}

func (eb *EditBuilder) Content(content string) *EditBuilder {
	eb.data.Content = option.NewNullableString(content)
	return eb
}

func (eb *EditBuilder) Embed(embed discord.Embed) *EditBuilder {
	eb.data.Embeds = &[]discord.Embed{embed}
	return eb
}

func (eb *EditBuilder) Embeds(embeds ...discord.Embed) *EditBuilder {
	eb.data.Embeds = &embeds
	return eb
}

func (eb *EditBuilder) ComponentsV2(components any) *EditBuilder {
	eb.Clear()
	eb.Flags(1 << 15)
	raw, _ := json.Marshal(components)

	comp, err := discord.ParseComponent(raw)
	if err != nil {
		panic(err)
	}

	cc := discord.ContainerComponents{comp.(discord.ContainerComponent)}
	eb.data.Components = &cc

	return eb
}

func (eb *EditBuilder) Components(components discord.ContainerComponents) *EditBuilder {
	eb.data.Components = &components
	return eb
}

func (eb *EditBuilder) Flags(flags ...discord.MessageFlags) *EditBuilder {
	for _, flag := range flags {
		eb.data.Flags |= flag
	}
	return eb
}

func (eb *EditBuilder) Clear() *EditBuilder {
	eb.data.Content = nil
	eb.data.Embeds = nil
	eb.data.Components = nil

	return eb
}

func (eb *EditBuilder) Send() (*discord.Message, error) {
	return eb.manager.state.EditInteractionResponse(eb.manager.appID, eb.manager.token, eb.data)
}
