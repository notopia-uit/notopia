package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermanentlyNewDeleteNoteHandler(t *testing.T) {
	t.Parallel()

	authSvc := &mockAuthorizationService{}
	noteRepo := &mockNoteRepo{}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, noteRepo, uow)

	require.NotNil(t, handler)
}

func TestPermanentlyDeleteNoteHandler_Handle_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteID := uuid.MustParse("aaaaaaaa-aaaa-bbbb-cccc-111111111111")
	workspaceID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	folderID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	userID := "user-del-note"

	note := domain.NewNote(noteID, "Delete Me", "", folderID)

	noteRepo := &mockNoteRepo{workspaceID: workspaceID, note: note}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}
	innerNoteRepo := &mockNoteRepo{note: note}
	innerRegistry := &mockRepoRegistry{noteRepo: innerNoteRepo}
	uow := &mockUnitOfWork{registry: innerRegistry}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, noteRepo, uow)

	cmd := &PermanentlyDeleteNote{
		ID:     noteID,
		UserID: userID,
	}

	err := handler.Handle(ctx, cmd)

	require.NoError(t, err)
	assert.True(t, uow.called, "unit of work should have been executed")
}

func TestPermanentlyDeleteNoteHandler_Handle_NoteRepoGetWorkspaceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	noteRepo := &mockNoteRepo{getWorkspaceErr: errSentinel}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, noteRepo, uow)

	cmd := &PermanentlyDeleteNote{
		ID:     noteID,
		UserID: "user-workspace-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
	assert.False(t, uow.called)
}

func TestPermanentlyDeleteNoteHandler_Handle_AuthServiceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	workspaceID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	noteRepo := &mockNoteRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{permissionErr: errSentinel}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, noteRepo, uow)

	cmd := &PermanentlyDeleteNote{
		ID:     noteID,
		UserID: "user-auth-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
	assert.False(t, uow.called)
}

func TestPermanentlyDeleteNoteHandler_Handle_NoPermission(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	workspaceID := uuid.MustParse("88888888-8888-8888-8888-888888888888")

	noteRepo := &mockNoteRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: false}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, noteRepo, uow)

	cmd := &PermanentlyDeleteNote{
		ID:     noteID,
		UserID: "user-no-perm",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)

	var forbidden *errs.Forbidden
	assert.ErrorAs(t, err, &forbidden, "should be a Forbidden error")
	assert.False(t, uow.called, "unit of work should not be called when permission denied")
}

func TestPermanentlyDeleteNoteHandler_Handle_UnitOfWorkGetByIDError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	workspaceID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	// inner note repo returns error when GetByID called inside uow
	innerNoteRepo := &mockNoteRepo{getByIDErr: errSentinel}
	innerRegistry := &mockRepoRegistry{noteRepo: innerNoteRepo}
	uow := &mockUnitOfWork{registry: innerRegistry}

	// outer note repo provides workspaceID
	outerNoteRepo := &mockNoteRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, outerNoteRepo, uow)

	cmd := &PermanentlyDeleteNote{
		ID:     noteID,
		UserID: "user-uow-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
}

func TestPermanentlyDeleteNoteHandler_Handle_UnitOfWorkSaveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	noteID := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	workspaceID := uuid.MustParse("abcdef12-abcd-abcd-abcd-abcdef123456")
	folderID := uuid.MustParse("fedcba98-fedc-fedc-fedc-fedcba987654")

	note := domain.NewNote(noteID, "To Delete", "", folderID)

	innerNoteRepo := &mockNoteRepo{note: note, saveErr: errSentinel}
	innerRegistry := &mockRepoRegistry{noteRepo: innerNoteRepo}
	uow := &mockUnitOfWork{registry: innerRegistry}

	outerNoteRepo := &mockNoteRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}

	handler := PermanentlyNewDeleteNoteHandler(authSvc, outerNoteRepo, uow)

	cmd := &PermanentlyDeleteNote{
		ID:     noteID,
		UserID: "user-save-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
}