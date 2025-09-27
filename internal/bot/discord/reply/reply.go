package reply

import (
	"encoding/json"
	"fmt"

	"github.com/nxtgo/arikawa/v3/api"
	"github.com/nxtgo/arikawa/v3/discord"
	"github.com/nxtgo/arikawa/v3/state"
	"github.com/nxtgo/arikawa/v3/utils/json/option"
)

type ResponseManager struct {
	state       *state.State
	interaction *discord.InteractionEvent
	token       string
	appID       discord.AppID
	deferred    bool
	responded   bool
}

type ResponseBuilder struct {
	manager *ResponseManager
	data    api.InteractionResponseData
}

func New(s *state.State, i *discord.InteractionEvent) *ResponseManager {
	return &ResponseManager{
		state:       s,
		interaction: i,
		token:       i.Token,
		appID:       i.AppID,
	}
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
		_, err := editBuilder.Clear().Embed(ErrorEmbed(err.Error())).Send()
		return err
	}

	return nil
}

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
