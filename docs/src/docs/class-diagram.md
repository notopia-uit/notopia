# Class Diagram

```plantuml
@startuml Notopia

skinparam classFontStyle bold
skinparam classAttributeIconSize 0
skinparam packageStyle rectangle
/' skinparam linetype ortho '/
/' skinparam linetype polyline '/

'Sytax of typescript'
package "Document" <<Bounded Context>> {
    entity "DocumentEntity" as Document.DocumentEntity {
        name: String
        data: Buffer
    }

    record "TagModel" as Document.TagModel {
        id: String
        name: String
    }

    record "LinkModel" as Document.LinkModel {
        documentId: String
        type: Backlink | OutgoingLink
    }

    class "DocumentModel" as Document.DocumentModel {
        name: String
        data: BlockNoteSchema[]
    }

    interface "DocumentRepository" as Document.DocumentRepository {
        Save(document Document)
        GetByID(documentId string): Document
    }

    record "AttachmentUploadUrl" as Document.AttachmentUploadUrl {
        id: string
        url: string
    }

    class "DocumentService" as Document.DocumentService {
        documentRepository: DocumentRepository
        blockNoteEditor: BlockNoteEditor
        atachmentService: AttachmentService

        getTags(): TagModel[]
        getLinks(): LinkModel[]
        CreateDocument(name string, data Buffer): Document
        GetDocument(documentId string): Document
        GetAttachmentUploadUrl(): AttachmentUploadUrl
    }

    DocumentService ..> DocumentRepository
}

'Syntax of golang'
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
                icon: *string
                workspaceID: uuid.UUID
                folderRelationship: FolderRelationship
                deletedBy: *DeletedBy
                deletedAt: *time.Time

                Rename(newName string)
                ParentID() *uuid.UUID, bool
                IsRoot() bool
                MoveToFolder(folderID uuid.UUID)
                Trash()
            }

            struct "FolderRelationship" as Note.Domain.Models.FolderRelationship <<Value Object>> {
                parentID: *uuid.UUID
                isRoot: bool

                ParentID() *uuid.UUID, bool
                IsRoot() bool
            }

            struct "Note" as Note.Domain.Models.Note <<Aggregate Root>> {
                id: uuid.UUID
                name: string
                icon: *string
                folderID: uuid.UUID
                tagIDs: []uuid.UUID
                outgoingLinks: []uuid.UUID
                backlinks: []uuid.UUID
                currentRevisionID: *uuid.UUID
                deletedBy: *DeletedBy
                deletedAt: *time.Time

                MoveNoteToFolder(folderID uuid.UUID)
                RemoveTag(tagID uuid.UUID)
                Trash()
                Restore()
            }

            struct "Tag" as Note.Domain.Models.Tag <<Aggregate Root>> {
                id: uuid.UUID
                name: string
                workspaceID: uuid.UUID
                stats: TagStats

                IncrementReference(delta int)
                DecrementReference(delta int)
                IsOrphaned() bool
            }

            struct "TagStats" as Note.Domain.Models.TagStats <<Value Object>> {
                referenceCount: int
                isExisting: bool

                IncrementReference(delta int) TagStats
                DecrementReference(delta int) TagStats
                ReferenceCount() int
                ShouldBePurged() bool
            }

            struct "Revision" as Note.Domain.Models.Revision <<Aggregate Root>> {
                id: uuid.UUID
                noteID: uuid.UUID
                name: string
                content: RevisionContent
                deletedAt: *time.Time

                rename(newName string)
            }

            struct "RevisionContent" as Note.Domain.Models.RevisionContent <<Value Object>> {
                blockNoteContent: string
                size: int

                BlockNoteContent() string
                Size() int
            }

            Note.Domain.Models.Workspace "1" *... "1..*" Note.Domain.Models.Folder : contains
            Note.Domain.Models.Folder "1" *-- "1" Note.Domain.Models.FolderRelationship : has
            Note.Domain.Models.Folder "1" *... "0..*" Note.Domain.Models.Note : contains
            Note.Domain.Models.Note "0..*" .. "0..*" Note.Domain.Models.Note : links
            Note.Domain.Models.Note "1..*" ...o "0..*" Note.Domain.Models.Tag : contains
            Note.Domain.Models.Tag "1" *-- "1" Note.Domain.Models.TagStats : has
            Note.Domain.Models.Note "1" *... "0..*" Note.Domain.Models.Revision : has
            Note.Domain.Models.Revision "1" *-- "1" Note.Domain.Models.RevisionContent : has
        }

        package "Repos" as Note.Domain.Repos {
            interface "Workspace" as Note.Domain.Repos.Workspace {
                GetByID(workspaceID uuid.UUID) *Workspace
                Save(workspace *Workspace)
            }

            interface "Folder" as Note.Domain.Repos.Folder {
                GetByID(folderID uuid.UUID) *Folder
                Save(folder *Folder)
                GetTrashedByWorkspaceID(workspaceID uuid.UUID, overDays *int) []*Folder
                PermanentlyDelete(folderIDs ...uuid.UUID)
            }

            interface "Note" as Note.Domain.Repos.Note {
                GetByID(noteID uuid.UUID) *Note
                Save(note *Note)
                GetTrashedByWorkspaceID(workspaceID uuid.UUID, overDays *int) []*Note
                PermanentlyDelete(noteIDs ...uuid.UUID)
            }

            interface "Tag" as Note.Domain.Repos.Tag {
                GetByID(tagID uuid.UUID) *Tag
                SearchByName(workspaceID uuid.UUID, name string) []*Tag
                Save(tag *Tag)
                Delete(tagID uuid.UUID)
            }

            interface "Revision" as Note.Domain.Repos.Revision {
                GetByID(revisionID uuid.UUID) *Revision
                Save(revision *Revision)
            }
        }

        package "Services" as Note.Domain.Services {
            struct "TagOrphan" as Note.Domain.Services.TagOrphan {
                noteRepo: Repos.Note
                tagRepo: Repos.Tag

                RemoveTagsFromNote(noteID uuid.UUID, tagIDs ...uuid.UUID)
            }
        }
    }
}
@enduml
```

<!-- diagram id="class-diagram" -->

<!-- vim:set tabstop=4 shiftwidth=4: -->
