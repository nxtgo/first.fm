package events

import (
	"reflect"

	"github.com/nxtgo/arikawa/v3/gateway"
	"go.fm/internal/bot/logging"
)

func TypeName(evt any) string {
	return reflect.TypeOf(evt).String()
}

// Handler is a generic event handler signature.
type Handler func(evt any)

// Registry holds all event handlers by event type.
type Registry struct {
	Handlers map[string][]Handler
	logger   func(name string) *logging.Logger
}

// NewRegistry creates a new event registry.
func NewRegistry() *Registry {
	return &Registry{
		Handlers: make(map[string][]Handler),
		logger: func(name string) *logging.Logger {
			return logging.WithFields(logging.F{"event_name": name})
		},
	}
}

// On registers a handler for a given event type.
func (r *Registry) On(eventName string, h Handler) {
	r.Handlers[eventName] = append(r.Handlers[eventName], h)
}

// Dispatch executes all handlers for a given event.
func (r *Registry) Dispatch(eventName string, evt any) {
	log := r.logger(eventName)
	for _, h := range r.Handlers[eventName] {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Errorw("panic in event handler", logging.F{"recover": rec})
				}
			}()
			h(evt)
		}()
	}
}

func RegisterDefaultEvents(r *Registry) {
	r.On(TypeName(&gateway.ReadyEvent{}), func(evt any) {
		if c, ok := evt.(*gateway.ReadyEvent); ok {
			r.logger("ready").Infow("client ready", logging.F{
				"tag":    c.User.Tag(),
				"guilds": len(c.Guilds),
			})
		}
	})
}
