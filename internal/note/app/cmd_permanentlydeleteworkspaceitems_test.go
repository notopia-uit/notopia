package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPermanentlyDeleteWorkspaceItemsHandler(t *testing.T) {
	t.Parallel()

	authSvc := &mockAuthorizationService{}
	uow := &mockUnitOfWork{}

	handler := NewPermanentlyDeleteWorkspaceItemsHandler(authSvc, uow)

	require.NotNil(t, handler)
}

func TestPermanentlyDeleteWorkspaceItemsHandler_Handle_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workspaceID := uuid.MustParse("aaaaaaaa-1111-aaaa-1111-aaaaaaaaaaaa")

	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}
	uow := &mockUnitOfWork{}

	handler := NewPermanentlyDeleteWorkspaceItemsHandler(authSvc, uow)

	cmd := &PermanentlyDeleteWorkspaceItems{
		WorkspaceID: workspaceID,
		UserID:      "user-ok",
		NoteIDs:     []uuid.UUID{uuid.MustParse("bbbbbbbb-2222-bbbb-2222-bbbbbbbbbbbb")},
		FolderIDs:   []uuid.UUID{uuid.MustParse("cccccccc-3333-cccc-3333-cccccccccccc")},
	}

	err := handler.Handle(ctx, cmd)

	// The impl is currently a stub that returns nil after auth check
	require.NoError(t, err)
}

func TestPermanentlyDeleteWorkspaceItemsHandler_Handle_NoPermission(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workspaceID := uuid.MustParse("dddddddd-4444-dddd-4444-dddddddddddd")

	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: false}
	uow := &mockUnitOfWork{}

	handler := NewPermanentlyDeleteWorkspaceItemsHandler(authSvc, uow)

	cmd := &PermanentlyDeleteWorkspaceItems{
		WorkspaceID: workspaceID,
		UserID:      "user-no-perm",
		NoteIDs:     []uuid.UUID{uuid.MustParse("eeeeeeee-5555-eeee-5555-eeeeeeeeeeee")},
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)

	var forbidden *errs.Forbidden
	assert.ErrorAs(t, err, &forbidden, "should return a Forbidden error")
}

func TestPermanentlyDeleteWorkspaceItemsHandler_Handle_AuthServiceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workspaceID := uuid.MustParse("ffffffff-6666-ffff-6666-ffffffffffff")

	authSvc := &mockAuthorizationService{permissionErr: errSentinel}
	uow := &mockUnitOfWork{}

	handler := NewPermanentlyDeleteWorkspaceItemsHandler(authSvc, uow)

	cmd := &PermanentlyDeleteWorkspaceItems{
		WorkspaceID: workspaceID,
		UserID:      "user-auth-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
}

func TestPermanentlyDeleteWorkspaceItemsHandler_Handle_EmptyLists(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workspaceID := uuid.MustParse("11111111-aaaa-1111-aaaa-111111111111")

	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}
	uow := &mockUnitOfWork{}

	handler := NewPermanentlyDeleteWorkspaceItemsHandler(authSvc, uow)

	cmd := &PermanentlyDeleteWorkspaceItems{
		WorkspaceID: workspaceID,
		UserID:      "user-empty",
		NoteIDs:     []uuid.UUID{},
		FolderIDs:   []uuid.UUID{},
	}

	err := handler.Handle(ctx, cmd)

	// Even with empty lists the stub returns nil
	require.NoError(t, err)
}