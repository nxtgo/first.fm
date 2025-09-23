package reply

import "github.com/nxtgo/arikawa/v3/api"

type DeferredResponse struct {
	manager *ResponseManager
	err     error
}

func (dr *DeferredResponse) Error() error {
	return dr.err
}

func (dr *DeferredResponse) Edit() *EditBuilder {
	return &EditBuilder{manager: dr.manager, data: api.EditInteractionResponseData{}}
}
