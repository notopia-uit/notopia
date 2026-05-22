package app

import "context"

type IdentitySvcActiveStatus uint8

const (
	IdentitySvcActiveStatusUnspecified IdentitySvcActiveStatus = iota
	IdentitySvcActiveStatusActive
	IdentitySvcActiveStatusInactive
)

type IdentitySvcSearchUsersParams struct {
	Keyword      string
	ActiveStatus IdentitySvcActiveStatus
	PaginationParams
}

type IdentitySvc interface {
	GetUsersByIDs(ctx context.Context, ids []string) ([]User, error)
	SearchUsers(ctx context.Context, params *IdentitySvcSearchUsersParams) (Paginated[User], error)
}
