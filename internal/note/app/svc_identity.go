package app

import "context"

type IdentityService interface {
	GetUsersByIDs(ctx context.Context, ids []string) ([]*User, error)
}
