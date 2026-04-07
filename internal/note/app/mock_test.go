package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

// ─────────────────────────────────────────────
// mockIntegrationPublisher
// ─────────────────────────────────────────────

type mockIntegrationPublisher struct {
	publishedEvents []IntegrationEvent
	err             error
}

func (m *mockIntegrationPublisher) Publish(_ context.Context, events ...IntegrationEvent) error {
	if m.err != nil {
		return m.err
	}
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

// ─────────────────────────────────────────────
// mockNoteRepo
// ─────────────────────────────────────────────

type mockNoteRepo struct {
	note          *domain.Note
	workspaceID   uuid.UUID
	getByIDErr    error
	getWorkspaceErr error
	saveErr       error
}

func (m *mockNoteRepo) GetByID(_ context.Context, _ uuid.UUID, _ bool) (*domain.Note, error) {
	return m.note, m.getByIDErr
}

func (m *mockNoteRepo) GetMany(_ context.Context, _ *domain.NoteRepoGetManyParams) ([]*domain.Note, error) {
	return nil, nil
}

func (m *mockNoteRepo) GetWorkspaceIDByID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return m.workspaceID, m.getWorkspaceErr
}

func (m *mockNoteRepo) Save(_ context.Context, _ *domain.Note) error {
	return m.saveErr
}

func (m *mockNoteRepo) SaveMany(_ context.Context, _ []*domain.Note) error {
	return nil
}

func (m *mockNoteRepo) AreAllInWorkspace(_ context.Context, _ []uuid.UUID, _ uuid.UUID) (bool, error) {
	return true, nil
}

// ─────────────────────────────────────────────
// mockFolderRepo
// ─────────────────────────────────────────────

type mockFolderRepo struct {
	folder          *domain.Folder
	workspaceID     uuid.UUID
	getByIDErr      error
	getWorkspaceErr error
	saveErr         error
}

func (m *mockFolderRepo) GetByID(_ context.Context, _ uuid.UUID, _ bool) (*domain.Folder, error) {
	return m.folder, m.getByIDErr
}

func (m *mockFolderRepo) GetMany(_ context.Context, _ *domain.FolderRepoGetManyParams) ([]*domain.Folder, error) {
	return nil, nil
}

func (m *mockFolderRepo) GetWorkspaceIDByID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return m.workspaceID, m.getWorkspaceErr
}

func (m *mockFolderRepo) Save(_ context.Context, _ *domain.Folder) error {
	return m.saveErr
}

func (m *mockFolderRepo) SaveMany(_ context.Context, _ []*domain.Folder) error {
	return nil
}

func (m *mockFolderRepo) AreAllInWorkspace(_ context.Context, _ []uuid.UUID, _ uuid.UUID) (bool, error) {
	return true, nil
}

// ─────────────────────────────────────────────
// mockRepoRegistry
// ─────────────────────────────────────────────

type mockRepoRegistry struct {
	noteRepo   domain.NoteRepo
	folderRepo domain.FolderRepo
}

func (m *mockRepoRegistry) Workspace() domain.WorkspaceRepo { return nil }
func (m *mockRepoRegistry) Folder() domain.FolderRepo       { return m.folderRepo }
func (m *mockRepoRegistry) Note() domain.NoteRepo           { return m.noteRepo }

// ─────────────────────────────────────────────
// mockUnitOfWork
// ─────────────────────────────────────────────

type mockUnitOfWork struct {
	registry domain.RepoRegistry
	err      error
	called   bool
}

func (m *mockUnitOfWork) Execute(ctx context.Context, fn func(domain.RepoRegistry) error) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return fn(m.registry)
}

// ─────────────────────────────────────────────
// mockAuthorizationService
// ─────────────────────────────────────────────

type mockAuthorizationService struct {
	hasWorkspaceItemPermission bool
	hasWorkspacePermission     bool
	permissionErr              error
}

func (m *mockAuthorizationService) HasWorkspacePermission(
	_ context.Context, _ string, _ uuid.UUID, _ WorkspacePermission,
) (bool, error) {
	return m.hasWorkspacePermission, m.permissionErr
}

func (m *mockAuthorizationService) HasWorkspaceItemPermission(
	_ context.Context, _ string, _ uuid.UUID, _ WorkspaceItemPermission,
) (bool, error) {
	return m.hasWorkspaceItemPermission, m.permissionErr
}

func (m *mockAuthorizationService) HasWorkspaceNotePermission(
	_ context.Context, _ string, _ uuid.UUID, _ WorkspaceItemPermission,
) (bool, error) {
	return m.hasWorkspaceItemPermission, m.permissionErr
}

func (m *mockAuthorizationService) HasWorkspaceFolderPermission(
	_ context.Context, _ string, _ uuid.UUID, _ WorkspaceItemPermission,
) (bool, error) {
	return m.hasWorkspaceItemPermission, m.permissionErr
}

func (m *mockAuthorizationService) CreateWorkspaceWithOwnership(
	_ context.Context, _ string, _ uuid.UUID, _ uuid.UUID,
) error {
	return nil
}

func (m *mockAuthorizationService) GetWorkspaceMembers(
	_ context.Context, _ string, _ uuid.UUID,
) ([]*WorkspaceMemberInfo, error) {
	return nil, nil
}

// ─────────────────────────────────────────────
// mockWorkspaceEventPublisher
// ─────────────────────────────────────────────

type mockWorkspaceEventPublisher struct {
	publishedEvents []WorkspaceEvent
	publishedIDs    []uuid.UUID
	err             error
	callCount       int
}

func (m *mockWorkspaceEventPublisher) Publish(
	_ context.Context,
	workspaceID uuid.UUID,
	_ string,
	events ...WorkspaceEvent,
) error {
	m.callCount++
	if m.err != nil {
		return m.err
	}
	m.publishedIDs = append(m.publishedIDs, workspaceID)
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

// ─────────────────────────────────────────────
// sentinelError
// ─────────────────────────────────────────────

var errSentinel = errors.New("sentinel error")