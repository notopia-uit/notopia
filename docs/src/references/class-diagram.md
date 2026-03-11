---
order: 3
---

# Class Diagram

## Note

```plantuml
@startuml Note
title Note

!$rosewater = "#dc8a78"
!$flamingo  = "#dd7878"
!$pink      = "#ea76cb"
!$mauve     = "#8839ef"
!$red       = "#d20f39"
!$maroon    = "#e64553"
!$peach     = "#fe640b"
!$yellow    = "#df8e1d"
!$green     = "#40a02b"
!$teal      = "#179299"
!$sky       = "#04a5e5"
!$sapphire  = "#209fb5"
!$lavender  = "#7287fd"
!$blue      = "#1e66f5"
!$text      = "#4c4f69"
!$subtext1  = "#5c5f77"
!$subtext0  = "#6c6f85"
!$overlay2  = "#7c7f93"
!$overlay1  = "#8c8fa1"
!$overlay0  = "#9ca0b0"
!$surface2  = "#acb0be"
!$surface1  = "#bcc0cc"
!$surface0  = "#ccd0da"
!$base      = "#eff1f5"
!$mantle    = "#e6e9ef"
!$crust     = "#dce0e8"

skinparam backgroundColor $base
skinparam defaultFontColor $text
skinparam roundcorner 16
skinparam classFontStyle bold

skinparam ArrowColor $subtext0
skinparam packageStyle rectangle

skinparam package {
    BackgroundColor $mantle
    BorderColor $surface1
    FontColor $mauve
}

skinparam class {
    BackgroundColor $base
    BorderColor $surface1
    HeaderBackgroundColor $surface0
    AttributeFontColor $text
}

!define RepoInterface(name) interface "name" as Domain.name <<(I, $pink) Repo Interface>>
!define ServiceInterface(name) interface "name" as Domain.name <<(I, $rosewater) Service Interface>>
!define Enum(name) enum "name" as Domain.name <<(E, $flamingo) Enum>>
!define ValueObject(name) class "name" as Domain.name <<(S, $yellow) Value Object>>
!define AggregateRoot(name) class "name" as Domain.name <<(S, $sky) Aggregate Root>>

package "Domain" as Domain <<Frame>> {
    Enum(TrashedBy) {
        PURPOSE
        PARENT
    }

    AggregateRoot(Workspace) {
        id: uuid.UUID
        name: string
        rootFolderID: uuid.UUID
        deletedAt: *time.Time

        Rename(newName string)
        Delete()
    }

    AggregateRoot(Folder) {
        id: uuid.UUID
        name: string
        icon: *string
        workspaceID: uuid.UUID
        folderRelationship: *FolderRelationship
        trashedBy: *TrashedBy
        deletedAt: *time.Time

        Rename(newName string)
        ParentID() *uuid.UUID, bool
        IsRoot() bool
        MoveToFolder(folderID uuid.UUID)
        Trash()
    }

    ValueObject(FolderRelationship) {
        parentID: *uuid.UUID
        isRoot: bool

        ParentID() *uuid.UUID, bool
        IsRoot() bool
    }

    AggregateRoot(Note) {
        id: uuid.UUID
        name: string
        icon: *string
        tags: []string
        size: uint
        folderID: uuid.UUID
        outgoingLinks: []uuid.UUID
        trashedBy: *TrashedBy
        deletedAt: *time.Time

        MoveNoteToFolder(folderID uuid.UUID)
        UpdateTags(tags []string)
        UpdateSize(size uint)
        Trash()
        Restore()
    }

    Domain.Workspace "1" *... "1..*" Domain.Folder : contains
    Domain.Folder "1" *-- "1" Domain.FolderRelationship : has
    Domain.Folder "1" *... "0..*" Domain.Note : contains
    Domain.Note "0..*" .. "0..*" Domain.Note : links

    RepoInterface(WorkspaceRepo) {
        CheckSlugExists(slug string) bool
        GetBySlug(slug string) *Workspace
        GetIDBySlug(slug string) *uuid.UUID
        Save(workspace *Workspace)
    }

    RepoInterface(FolderRepo) {
        GetByID(folderID uuid.UUID) *Folder
        Save(folder *Folder)
        GetTrashedByWorkspaceID(workspaceID uuid.UUID, overDays *int) []Folder
        PermanentlyDelete(folderIDs ...uuid.UUID)
    }

    RepoInterface(NoteRepo) {
        GetByID(noteID uuid.UUID) *Note
        Save(note *Note)
        GetTrashedByWorkspaceID(workspaceID uuid.UUID, overDays *int) []Note
        PermanentlyDelete(noteIDs ...uuid.UUID)
    }
}
@enduml
```

<!-- diagram id="class-diagram-note" -->

:::info

- Golang syntax
- Apply DDD, CQRS, repository pattern

:::

## Document

```plantuml
@startuml Document
title Document

!$rosewater = "#dc8a78"
!$flamingo  = "#dd7878"
!$pink      = "#ea76cb"
!$mauve     = "#8839ef"
!$red       = "#d20f39"
!$maroon    = "#e64553"
!$peach     = "#fe640b"
!$yellow    = "#df8e1d"
!$green     = "#40a02b"
!$teal      = "#179299"
!$sky       = "#04a5e5"
!$sapphire  = "#209fb5"
!$lavender  = "#7287fd"
!$blue      = "#1e66f5"
!$text      = "#4c4f69"
!$subtext1  = "#5c5f77"
!$subtext0  = "#6c6f85"
!$overlay2  = "#7c7f93"
!$overlay1  = "#8c8fa1"
!$overlay0  = "#9ca0b0"
!$surface2  = "#acb0be"
!$surface1  = "#bcc0cc"
!$surface0  = "#ccd0da"
!$base      = "#eff1f5"
!$mantle    = "#e6e9ef"
!$crust     = "#dce0e8"

skinparam backgroundColor $base
skinparam defaultFontColor $text
skinparam roundcorner 16
skinparam classFontStyle bold
skinparam ArrowColor $subtext0
skinparam packageStyle rectangle

skinparam package {
    BackgroundColor $mantle
    BorderColor $surface1
    FontColor $mauve
}

skinparam class {
    BackgroundColor $base
    BorderColor $surface1
    HeaderBackgroundColor $surface0
    AttributeFontColor $text
}

!define RepoInterface(name) interface "name" as Document.name <<(I, $pink) Repo Interface>>
!define Service(name) class "name" as Document.name <<(C, $rosewater) Service>>
!define Entity(name) class "name" as Document.name <<(C, $sky) Entity>>
!define Type(name) class "name" as Document.name <<(T, $flamingo) Type>>
!define Model(name) class "name" as Document.name <<(C, $yellow) Model>>

package "Document" as Document <<Frame>> {
    Entity(DocumentEntity) {
        id: string
        data: Buffer
    }

    Entity(RevisionEntity) {
        id: string
        documentId: string
        data: Buffer
    }

    Model(DocumentModel) {
        id: string
        content: @blocknote/core.Block[]
    }

    Model(RevisionModel) {
        id: string
        documentId: string
        content: @blocknote/core.Block[]
    }

    Type(AttachmentUploadUrl) {
        id: string
        url: string
    }

    Type(TagModel) {
        id: string
        name: string
    }

    RepoInterface(DocumentRepository) {
        Save(document: DocumentEntity)
        GetById(documentId: string): DocumentEntity
    }

    RepoInterface(RevisionRepository) {
        Save(revision: RevisionEntity)
        GetById(revisionId: string): RevisionEntity
        GetByDocumentId(documentId: string): RevisionEntity[]
    }

    Service(DocumentService) {
        documentRepository: DocumentRepository
        blockNoteEditor: BlockNoteEditor
        attachmentService: AttachmentService

        extractTags(): TagModel[]
        extractOutgoingLinkIds(): string[]
        CreateDocument(name: string, data: Buffer): DocumentEntity
        GetDocument(documentId: string): DocumentEntity
        GetAttachmentUploadUrl(): AttachmentUploadUrl
    }

    Service(RevisionService) {
        revisionRepository: RevisionRepository

        GetRevision(revisionId: string): RevisionModel
        GetRevisionsByDocumentId(documentId: string): RevisionModel[]
    }

    Document.DocumentService ..> Document.DocumentRepository : uses
    Document.RevisionService ..> Document.RevisionRepository : uses
    Document.DocumentRepository ..> Document.DocumentEntity : manages
    Document.RevisionRepository ..> Document.RevisionEntity : manages
}

@endum
```

<!-- diagram id="class-diagram-document" -->

:::info

- Typescript syntax
- Apply layered architecture, repository pattern

:::

## Authorization

```plantuml
@startuml Document
title Document

!$rosewater = "#dc8a78"
!$flamingo  = "#dd7878"
!$pink      = "#ea76cb"
!$mauve     = "#8839ef"
!$red       = "#d20f39"
!$maroon    = "#e64553"
!$peach     = "#fe640b"
!$yellow    = "#df8e1d"
!$green     = "#40a02b"
!$teal      = "#179299"
!$sky       = "#04a5e5"
!$sapphire  = "#209fb5"
!$lavender  = "#7287fd"
!$blue      = "#1e66f5"
!$text      = "#4c4f69"
!$subtext1  = "#5c5f77"
!$subtext0  = "#6c6f85"
!$overlay2  = "#7c7f93"
!$overlay1  = "#8c8fa1"
!$overlay0  = "#9ca0b0"
!$surface2  = "#acb0be"
!$surface1  = "#bcc0cc"
!$surface0  = "#ccd0da"
!$base      = "#eff1f5"
!$mantle    = "#e6e9ef"
!$crust     = "#dce0e8"

skinparam backgroundColor $base
skinparam defaultFontColor $text
skinparam roundcorner 16
skinparam classFontStyle bold
skinparam ArrowColor $subtext0
skinparam packageStyle rectangle

skinparam package {
    BackgroundColor $mantle
    BorderColor $surface1
    FontColor $mauve
}

skinparam class {
    BackgroundColor $base
    BorderColor $surface1
    HeaderBackgroundColor $surface0
    AttributeFontColor $text
}

!define Enum(name) enum "name" as name <<(E, $flamingo) Enum>>

Enum(WorkspaceRole) {
    OWNER
    EDITOR
    VIEWER
}
```

<!-- diagram id="class-diagram-authorization" -->

:::info

- Golang syntax

:::

<!-- vim:set tabstop=4 shiftwidth=4: -->
