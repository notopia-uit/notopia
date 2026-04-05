package domainevent

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

type adapter[E domain.Event] struct {
	handler app.DomainEventHandler[E]
}

var _ cqrs.EventHandler = (*adapter[domain.Event])(nil)

func (a *adapter[E]) HandlerName() string {
	return fmt.Sprintf("%T", a.handler)
}

func (a *adapter[E]) NewEvent() any {
	var event E
	eventType := reflect.TypeOf(event)

	if eventType.Kind() == reflect.Ptr {
		return reflect.New(eventType.Elem()).Interface()
	}
	return reflect.New(eventType).Interface()
}

func (a *adapter[E]) Handle(ctx context.Context, event any) error {
	typedEvent, ok := event.(E)
	if !ok {
		return fmt.Errorf("invalid event type: expected %T, got %T", new(E), event)
	}
	return a.handler.Handle(ctx, typedEvent)
}
