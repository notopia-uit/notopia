package app

import (
	"context"
)

type IntegrationPub interface {
	Publish(ctx context.Context, event any) error
}
