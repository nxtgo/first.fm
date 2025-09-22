package reply

import (
	"fmt"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/utils/json"
	"github.com/nxtgo/arikawa/v3/utils/json/option"
)

type ResponseBuilder struct {
	manager *ResponseManager
	data    api.InteractionResponseData
}

func (rm *ResponseManager) Reply() *ResponseBuilder {
	return &ResponseBuilder{
		manager: rm,
		data:    api.InteractionResponseData{},
	}
}

func (rm *ResponseManager) Defer(flags ...discord.MessageFlags) *DeferredResponse {
	if rm.responded {
		return &DeferredResponse{manager: rm, err: fmt.Errorf("already responded")}
	}

	var combinedFlags discord.MessageFlags
	for _, flag := range flags {
		combinedFlags |= flag
	}

	response := api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{Flags: combinedFlags},
	}

	err := rm.state.RespondInteraction(rm.interaction.ID, rm.token, response)
	rm.deferred = true
	rm.responded = true

	return &DeferredResponse{manager: rm, err: err}
}

// no

func (rb *ResponseBuilder) Content(content string) *ResponseBuilder {
	rb.data.Content = option.NewNullableString(content)
	return rb
}

func (rb *ResponseBuilder) Embed(embed discord.Embed) *ResponseBuilder {
	rb.data.Embeds = &[]discord.Embed{embed}
	return rb
}

func (rb *ResponseBuilder) ComponentsV2(components any) *ResponseBuilder {
	rb.Flags(1 << 15)
	raw, _ := json.Marshal(components)

	comp, err := discord.ParseComponent(raw)
	if err != nil {
		panic(err)
	}

	cc := discord.ContainerComponents{comp.(discord.ContainerComponent)}
	rb.data.Components = &cc

	return rb
}

func (rb *ResponseBuilder) Components(components discord.ContainerComponents) *ResponseBuilder {
	rb.data.Components = &components
	return rb
}

func (rb *ResponseBuilder) Flags(flags ...discord.MessageFlags) *ResponseBuilder {
	for _, flag := range flags {
		rb.data.Flags |= flag
	}
	return rb
}

func (rb *ResponseBuilder) Send() error {
	if rb.manager.responded {
		return fmt.Errorf("already responded")
	}

	err := rb.manager.state.RespondInteraction(
		rb.manager.interaction.ID,
		rb.manager.token,
		api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &rb.data,
		},
	)

	rb.manager.responded = true
	return err
}
