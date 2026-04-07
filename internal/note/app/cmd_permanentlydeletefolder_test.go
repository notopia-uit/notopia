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

func TestPermanentlyNewDeleteFolderHandler(t *testing.T) {
	t.Parallel()

	authSvc := &mockAuthorizationService{}
	folderRepo := &mockFolderRepo{}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteFolderHandler(authSvc, folderRepo, uow)

	require.NotNil(t, handler)
}

func TestPermanentlyDeleteFolderHandler_Handle_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	folderID := uuid.MustParse("aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa")
	workspaceID := uuid.MustParse("bbbbbbbb-2222-2222-2222-bbbbbbbbbbbb")
	userID := "user-ok"

	folder, err := domain.NewFolder(folderID, "Test Folder", "", workspaceID, domain.FolderHierarchy{}, userID)
	require.NoError(t, err)

	folderRepo := &mockFolderRepo{workspaceID: workspaceID, folder: folder}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}
	innerRepo := &mockRepoRegistry{folderRepo: folderRepo}
	uow := &mockUnitOfWork{registry: innerRepo}

	handler := PermanentlyNewDeleteFolderHandler(authSvc, folderRepo, uow)

	cmd := &PermanentlyDeleteFolder{
		ID:     folderID,
		UserID: userID,
	}

	handleErr := handler.Handle(ctx, cmd)

	require.NoError(t, handleErr)
	assert.True(t, uow.called, "unit of work should have been executed")
}

func TestPermanentlyDeleteFolderHandler_Handle_FolderRepoGetWorkspaceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	folderID := uuid.MustParse("cccccccc-3333-3333-3333-cccccccccccc")

	folderRepo := &mockFolderRepo{getWorkspaceErr: errSentinel}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteFolderHandler(authSvc, folderRepo, uow)

	cmd := &PermanentlyDeleteFolder{
		ID:     folderID,
		UserID: "user-bad",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
	assert.False(t, uow.called, "unit of work should not have been called")
}

func TestPermanentlyDeleteFolderHandler_Handle_AuthServiceError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	folderID := uuid.MustParse("dddddddd-4444-4444-4444-dddddddddddd")
	workspaceID := uuid.MustParse("eeeeeeee-5555-5555-5555-eeeeeeeeeeee")

	folderRepo := &mockFolderRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{permissionErr: errSentinel}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteFolderHandler(authSvc, folderRepo, uow)

	cmd := &PermanentlyDeleteFolder{
		ID:     folderID,
		UserID: "user-auth-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
	assert.False(t, uow.called)
}

func TestPermanentlyDeleteFolderHandler_Handle_NoPermission(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	folderID := uuid.MustParse("ffffffff-6666-6666-6666-ffffffffffff")
	workspaceID := uuid.MustParse("11111111-7777-7777-7777-111111111111")

	folderRepo := &mockFolderRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: false}
	uow := &mockUnitOfWork{}

	handler := PermanentlyNewDeleteFolderHandler(authSvc, folderRepo, uow)

	cmd := &PermanentlyDeleteFolder{
		ID:     folderID,
		UserID: "user-no-perm",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)

	var forbidden *errs.Forbidden
	assert.ErrorAs(t, err, &forbidden, "error should be a Forbidden error")
	assert.False(t, uow.called, "unit of work should not be called when permission denied")
}

func TestPermanentlyDeleteFolderHandler_Handle_UnitOfWorkGetByIDError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	folderID := uuid.MustParse("22222222-8888-8888-8888-222222222222")
	workspaceID := uuid.MustParse("33333333-9999-9999-9999-333333333333")

	// folderRepo returns an error when GetByID is called inside uow
	innerFolderRepo := &mockFolderRepo{getByIDErr: errSentinel}
	innerRegistry := &mockRepoRegistry{folderRepo: innerFolderRepo}
	uow := &mockUnitOfWork{registry: innerRegistry}

	// outer folderRepo provides workspaceID without error
	outerFolderRepo := &mockFolderRepo{workspaceID: workspaceID}
	authSvc := &mockAuthorizationService{hasWorkspaceItemPermission: true}

	handler := PermanentlyNewDeleteFolderHandler(authSvc, outerFolderRepo, uow)

	cmd := &PermanentlyDeleteFolder{
		ID:     folderID,
		UserID: "user-uow-err",
	}

	err := handler.Handle(ctx, cmd)

	require.Error(t, err)
	assert.ErrorIs(t, err, errSentinel)
}