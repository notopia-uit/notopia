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

        struct "Workspace" as Note.Domain.Workspace <<Aggregate Root>> {
            id: uuid.UUID
            name: string
            rootFolderID: uuid.UUID
            deletedAt: *time.Time

            Rename(newName string)
            Delete()
        }

        struct "Folder" as Note.Domain.Folder <<Aggregate Root>> {
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

        struct "FolderRelationship" as Note.Domain.FolderRelationship <<Value Object>> {
            parentID: *uuid.UUID
            isRoot: bool

            ParentID() *uuid.UUID, bool
            IsRoot() bool
        }

        struct "Note" as Note.Domain.Note <<Aggregate Root>> {
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

        struct "Tag" as Note.Domain.Tag <<Aggregate Root>> {
            id: uuid.UUID
            name: string
            workspaceID: uuid.UUID
            stats: TagStats

            IncrementReference(delta int)
            DecrementReference(delta int)
            IsOrphaned() bool
        }

        struct "TagStats" as Note.Domain.TagStats <<Value Object>> {
            referenceCount: int
            isExisting: bool

            IncrementReference(delta int) TagStats
            DecrementReference(delta int) TagStats
            ReferenceCount() int
            ShouldBePurged() bool
        }

        struct "Revision" as Note.Domain.Revision <<Aggregate Root>> {
            id: uuid.UUID
            noteID: uuid.UUID
            name: string
            content: RevisionContent
            deletedAt: *time.Time

            rename(newName string)
        }

        struct "RevisionContent" as Note.Domain.RevisionContent <<Value Object>> {
            blockNoteContent: string
            size: int

            BlockNoteContent() string
            Size() int
        }

        Note.Domain.Workspace "1" *... "1..*" Note.Domain.Folder : contains
        Note.Domain.Folder "1" *-- "1" Note.Domain.FolderRelationship : has
        Note.Domain.Folder "1" *... "0..*" Note.Domain.Note : contains
        Note.Domain.Note "0..*" .. "0..*" Note.Domain.Note : links
        Note.Domain.Note "1..*" ...o "0..*" Note.Domain.Tag : contains
        Note.Domain.Tag "1" *-- "1" Note.Domain.TagStats : has
        Note.Domain.Note "1" *... "0..*" Note.Domain.Revision : has
        Note.Domain.Revision "1" *-- "1" Note.Domain.RevisionContent : has

        interface "WorkspaceRepo" as Note.Domain.WorkspaceRepo {
            GetByID(workspaceID uuid.UUID) *Workspace
            Save(workspace *Workspace)
        }

        interface "FolderRepo" as Note.Domain.FolderRepo {
            GetByID(folderID uuid.UUID) *Folder
            Save(folder *Folder)
            GetTrashedByWorkspaceID(workspaceID uuid.UUID, overDays *int) []*Folder
            PermanentlyDelete(folderIDs ...uuid.UUID)
        }

        interface "NoteRepo" as Note.Domain.NoteRepo {
            GetByID(noteID uuid.UUID) *Note
            Save(note *Note)
            GetTrashedByWorkspaceID(workspaceID uuid.UUID, overDays *int) []*Note
            PermanentlyDelete(noteIDs ...uuid.UUID)
        }

        interface "TagRepo" as Note.Domain.TagRepo {
            GetByID(tagID uuid.UUID) *Tag
            SearchByName(workspaceID uuid.UUID, name string) []*Tag
            Save(tag *Tag)
            Delete(tagID uuid.UUID)
        }

        interface "RevisionRepo" as Note.Domain.RevisionRepo {
            GetByID(revisionID uuid.UUID) *Revision
            Save(revision *Revision)
        }

        struct "TagOrphanService" as Note.Domain.TagOrphanService {
            noteRepo: Repos.Note
            tagRepo: Repos.Tag

            RemoveTagsFromNote(noteID uuid.UUID, tagIDs ...uuid.UUID)
        }
    }
}
@enduml
```

<!-- diagram id="class-diagram" -->

<!-- vim:set tabstop=4 shiftwidth=4: -->
