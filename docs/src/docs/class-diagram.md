# Class Diagram

```plantuml
@startuml Notopia

skinparam classFontStyle bold
skinparam classAttributeIconSize 0
skinparam packageStyle rectangle
/' skinparam linetype ortho '/
/' skinparam linetype polyline '/

package "Document" {
    class "Document" as Document.Document {
        name: string
        data: []byte
    }
}

package "Note" <<Bounded Context>> {
    package "Domain" as Note.Domain <<Frame>> {
        enum "WorkspaceRole" as Note.Domain.WorkspaceRole {
            OWNER
            EDITOR
            VIEWER
        }

        enum "DeletedBy" as Note.Domain.DeletedBy {
            PURPOSE
            PARENT
        }

        package "Models" as Note.Domain.Models {
            struct "Workspace" as Note.Domain.Models.Workspace <<Aggregate Root>> {
                id: uuid.UUID
                name: string
                rootFolderID: uuid.UUID
                deletedAt: *time.Time

                Rename(newName string)
                Delete()
            }

            struct "Folder" as Note.Domain.Models.Folder <<Aggregate Root>> {
                id: uuid.UUID
                name: string
                parent: *Folder
                deletedBy: *DeletedBy
                deletedAt: *time.Time

                Rename(newName string)
                MoveToFolder(folderID uuid.UUID)
                Trash()
            }

            struct "Note" as Note.Domain.Models.Note <<Aggregate Root>> {
                id: uuid.UUID
                name: string
                FolderID: uuid.UUID
                tags: []string
                currentRevisionID: uuid.UUID
                deletedBy: *DeletedBy
                deletedAt: *time.Time

                MoveNoteToFolder(folderID uuid.UUID)
                ReplaceTags(tags []string)
                Trash()
                Restore()
            }

            struct "Revision" as Note.Domain.Models.Revision <<Aggregate Root>> {
                id: uuid.UUID
                noteID: uuid.UUID
                name: string
                blockNoteContent: string
                deletedAt: *time.Time

                rename(newName)
            }

        }

        package "Repos" as Note.Domain.Repos <<Frame>> {
            interface "Workspace" as Note.Domain.Repos.Workspace {
                GetByID(workspaceID uuid.UUID) *Workspace
                Save(workspace *Workspace)
            }

            package "Folder" as Note.Domain.Repos.Folder {
                struct "Options" as Note.Domain.Repos.Folder.Options {
                    OverDays: *int
                }

                interface "Folder" as Note.Domain.Repos.Folder.Folder {
                    GetByID(folderID uuid.UUID) *Folder
                    Save(folder *Folder)
                    GetTrashedByWorkspaceID(workspaceID uuid.UUID, options *Options) []*Folder
                }
            }

            package "Note" as Note.Domain.Repos.Note {
                struct "Options" as Note.Domain.Repos.Note.Options {
                    OverDays: *int
                }

                interface "Note" as Note.Domain.Repos.Note.Note {
                    GetByID(noteID uuid.UUID) *Note
                    Save(note *Note)
                    GetTrashedByWorkspaceID(workspaceID uuid.UUID, options *Options) []*Note
                }
            }

            interface "Revision" as Note.Domain.Repos.Revision {
                GetByID(revisionID uuid.UUID) *Revision
                Save(revision *Revision)
            }
        }
    }
}
@enduml
```

<!-- diagram id="class-diagram" -->

<!-- vim:set tabstop=4 shiftwidth=4: -->
