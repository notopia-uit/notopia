package app

import "context"

type IdentitySvc interface {
	GetUsersByIDs(ctx context.Context, ids []string) ([]*User, error)
}
