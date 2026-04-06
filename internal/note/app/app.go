package app

type Cmds struct {
	CreateFolderHandler                    *CreateFolderHandler
	CreateNoteHandler                      *CreateNoteHandler
	CreateWorkspaceHandler                 *CreateWorkspaceHandler
	DeleteFolderHandler                    *PermanentlyDeleteFolderHandler
	DeleteNoteHandler                      *PermanentlyDeleteNoteHandler
	DeleteWorkspaceHandler                 *DeleteWorkspaceHandler
	GenerateDailyNoteHandler               *GenerateDailyNoteHandler
	MoveWorkspaceItemsHandler              *MoveWorkspaceItemsHandler
	PermanentlyDeleteWorkspaceItemsHandler *PermanentlyDeleteWorkspaceItemsHandler
	PublishNoteHandler                     *PublishNoteHandler
	PublishWorkspaceHandler                *PublishWorkspaceHandler
	RenameFolderHandler                    *RenameFolderHandler
	RenameNoteHandler                      *RenameNoteHandler
	RenameWorkspaceHandler                 *RenameWorkspaceHandler
	RestoreTrashedWorkspaceItemsHandler    *RestoreTrashedWorkspaceItemsHandler
	TrashWorkspaceItemsHandler             *TrashWorkspaceItemsHandler
	UnpublishNoteHandler                   *UnpublishNoteHandler
	UnpublishWorkspaceHandler              *UnpublishWorkspaceHandler
	UpdateWorkspaceMembersHandler          *UpdateWorkspaceMembersHandler
}

type Queries struct {
	CheckWorkspaceSlugExistsHandler *CheckWorkspaceSlugExistsHandler
	GetNoteGraphHandler             *GetNoteGraphHandler
	GetNoteHandler                  *GetNoteHandler
	GetNoteLinksHandler             *GetNoteLinksHandler
	GetWorkspaceGraphHandler        *GetWorkspaceGraphHandler
	GetWorkspaceHandler             *GetWorkspaceHandler
	GetWorkspaceMembersHandler      *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler         *GetWorkspaceTreeHandler
	ShowTrashHandler                *ShowTrashHandler
}

type IntegrationEvents struct {
	DocumentCommittedHandler *DocumentCommittedHandler
}

type Server struct {
	Cmds              *Cmds
	IntegrationEvents *IntegrationEvents
	Queries           *Queries

	WorkspaceEventHub WorkspaceEventHub
}
