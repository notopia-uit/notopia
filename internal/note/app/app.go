package app

type App struct {
	// Command Handlers
	CreateNoteHandler                   *CreateNoteHandler
	CreateFolderHandler                 *CreateFolderHandler
	CreateWorkspaceHandler              *CreateWorkspaceHandler
	DeleteNoteHandler                   *DeleteNoteHandler
	DeleteWorkspaceHandler              *DeleteWorkspaceHandler
	GenerateDailyNoteHandler            *GenerateDailyNoteHandler
	MoveWorkspaceItemsHandler           *MoveWorkspaceItemsHandler
	PublishNoteHandler                  *PublishNoteHandler
	PublishWorkspaceHandler             *PublishWorkspaceHandler
	RenameFolderHandler                 *RenameFolderHandler
	RenameNoteHandler                   *RenameNoteHandler
	RenameWorkspaceHandler              *RenameWorkspaceHandler
	RestoreTrashedWorkspaceItemsHandler *RestoreTrashedWorkspaceItemsHandler
	TrashWorkspaceItemsHandler          *TrashWorkspaceItemsHandler
	UnpublishNoteHandler                *UnpublishNoteHandler
	UnpublishWorkspaceHandler           *UnpublishWorkspaceHandler
	UpdateWorkspaceMembersHandler       *UpdateWorkspaceMembersHandler

	// Query Handlers
	CheckWorkspaceExistsHandler *CheckWorkspaceExistsHandler
	GetNoteGraphHandler         *GetNoteGraphHandler
	GetNoteLinksHandler         *GetNoteLinksHandler
	GetNotesHandler             *GetNotesHandler
	GetWorkspaceHandler         *GetWorkspaceHandler
	GetWorkspaceGraphHandler    *GetWorkspaceGraphHandler
	GetWorkspaceMembersHandler  *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler     *GetWorkspaceTreeHandler
	ShowTrashHandler            *ShowTrashHandler

	// Event Handlers
	DocumentCommittedHandler *DocumentCommittedHandler
}

func NewApp(
	createNoteHandler *CreateNoteHandler,
	createFolderHandler *CreateFolderHandler,
	createWorkspaceHandler *CreateWorkspaceHandler,
	deleteNoteHandler *DeleteNoteHandler,
	deleteWorkspaceHandler *DeleteWorkspaceHandler,
	generateDailyNoteHandler *GenerateDailyNoteHandler,
	moveWorkspaceItemsHandler *MoveWorkspaceItemsHandler,
	publishNoteHandler *PublishNoteHandler,
	publishWorkspaceHandler *PublishWorkspaceHandler,
	renameFolderHandler *RenameFolderHandler,
	renameNoteHandler *RenameNoteHandler,
	renameWorkspaceHandler *RenameWorkspaceHandler,
	restoreTrashedWorkspaceItemsHandler *RestoreTrashedWorkspaceItemsHandler,
	trashWorkspaceItemsHandler *TrashWorkspaceItemsHandler,
	unpublishNoteHandler *UnpublishNoteHandler,
	unpublishWorkspaceHandler *UnpublishWorkspaceHandler,
	updateWorkspaceMembersHandler *UpdateWorkspaceMembersHandler,
	checkWorkspaceExistsHandler *CheckWorkspaceExistsHandler,
	getNoteGraphHandler *GetNoteGraphHandler,
	getNoteLinksHandler *GetNoteLinksHandler,
	getNotesHandler *GetNotesHandler,
	getWorkspaceHandler *GetWorkspaceHandler,
	getWorkspaceGraphHandler *GetWorkspaceGraphHandler,
	getWorkspaceMembersHandler *GetWorkspaceMembersHandler,
	getWorkspaceTreeHandler *GetWorkspaceTreeHandler,
	showTrashHandler *ShowTrashHandler,
	documentCommittedHandler *DocumentCommittedHandler,
) *App {
	return &App{
		CreateNoteHandler:                   createNoteHandler,
		CreateFolderHandler:                 createFolderHandler,
		CreateWorkspaceHandler:              createWorkspaceHandler,
		DeleteNoteHandler:                   deleteNoteHandler,
		DeleteWorkspaceHandler:              deleteWorkspaceHandler,
		GenerateDailyNoteHandler:            generateDailyNoteHandler,
		MoveWorkspaceItemsHandler:           moveWorkspaceItemsHandler,
		PublishNoteHandler:                  publishNoteHandler,
		PublishWorkspaceHandler:             publishWorkspaceHandler,
		RenameFolderHandler:                 renameFolderHandler,
		RenameNoteHandler:                   renameNoteHandler,
		RenameWorkspaceHandler:              renameWorkspaceHandler,
		RestoreTrashedWorkspaceItemsHandler: restoreTrashedWorkspaceItemsHandler,
		TrashWorkspaceItemsHandler:          trashWorkspaceItemsHandler,
		UnpublishNoteHandler:                unpublishNoteHandler,
		UnpublishWorkspaceHandler:           unpublishWorkspaceHandler,
		UpdateWorkspaceMembersHandler:       updateWorkspaceMembersHandler,
		CheckWorkspaceExistsHandler:         checkWorkspaceExistsHandler,
		GetNoteGraphHandler:                 getNoteGraphHandler,
		GetNoteLinksHandler:                 getNoteLinksHandler,
		GetNotesHandler:                     getNotesHandler,
		GetWorkspaceHandler:                 getWorkspaceHandler,
		GetWorkspaceGraphHandler:            getWorkspaceGraphHandler,
		GetWorkspaceMembersHandler:          getWorkspaceMembersHandler,
		GetWorkspaceTreeHandler:             getWorkspaceTreeHandler,
		ShowTrashHandler:                    showTrashHandler,
		DocumentCommittedHandler:            documentCommittedHandler,
	}
}

var ProvideApp = NewApp
