package event

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
)

type PubSub struct {
	commandBus       *cqrs.CommandBus
	commandProcessor *cqrs.CommandProcessor
	eventBus         *cqrs.EventBus
	eventProcessor   *cqrs.EventProcessor
	router           *message.Router
	app              *app.App
}

func NewPubSub(
	commandBus *cqrs.CommandBus,
	commandProcessor *cqrs.CommandProcessor,
	eventBus *cqrs.EventBus,
	eventProcessor *cqrs.EventProcessor,
	router *message.Router,
	app *app.App,
) *PubSub {
	return &PubSub{
		commandBus:       commandBus,
		commandProcessor: commandProcessor,
		eventBus:         eventBus,
		eventProcessor:   eventProcessor,
		router:           router,
		app:              app,
	}
}
