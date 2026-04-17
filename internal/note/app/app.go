package app

type Cmds struct {
	ChangeWorkspaceSlugHandler             *ChangeWorkspaceSlugHandler
	CreateFolderHandler                    *CreateFolderHandler
	CreateNoteHandler                      *CreateNoteHandler
	CreateWorkspaceHandler                 *CreateWorkspaceHandler
	DeleteFolderHandler                    *PermanentlyDeleteFolderHandler
	DeleteNoteHandler                      *PermanentlyDeleteNoteHandler
	DeleteWorkspaceHandler                 *DeleteWorkspaceHandler
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
	GetMyWorkspacesHandler          *GetMyWorkspacesHandler
	GetNoteGraphHandler             *GetNoteGraphHandler
	GetNoteHandler                  *GetNoteHandler
	GetWorkspaceByNoteHandler       *GetWorkspaceByNoteHandler
	GetNoteLinksHandler             *GetNoteLinksHandler
	GetWorkspaceGraphHandler        *GetWorkspaceGraphHandler
	GetWorkspaceHandler             *GetWorkspaceHandler
	GetWorkspaceMembersHandler      *GetWorkspaceMembersHandler
	GetWorkspaceTreeHandler         *GetWorkspaceTreeHandler
	ShowTrashHandler                *ShowTrashHandler
}

type Events struct {
	DocumentCommittedHandler           *DocumentCommittedHandler
	NotifyWorkspaceItemsUpdatedHandler *NotifyWorkspaceItemsUpdatedHandler
	NotifyWorkspaceRenamedHandler      *NotifyWorkspaceRenamedHandler
	NotifyWorkspaceSlugChangedHandler  *NotifyWorkspaceSlugChangedHandler
}

type Server struct {
	Cmds    *Cmds
	Events  *Events
	Queries *Queries

	WorkspaceEventHub WorkspaceEventHub
}
