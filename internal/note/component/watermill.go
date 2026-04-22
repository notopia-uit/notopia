package component

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func NewWatermillJsonMarshaler() *cqrs.JSONMarshaler {
	return &cqrs.JSONMarshaler{}
}

var ProvideWatermillJsonMarshaler = NewWatermillJsonMarshaler
