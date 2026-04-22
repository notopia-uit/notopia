---
order: 1
---

# Document Class Diagram

:::info

- Typescript syntax
- Apply layered architecture, use NestJS framework

:::

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

!define Service(name) class "name" as Document.name <<(C, $rosewater) Service>>
!define Entity(name) class "name" as Document.name <<(C, $sky) Entity>>
!define Type(name) class "name" as Document.name <<(T, $flamingo) Type>>
!define Enum(name) enum "name" as Document.name <<(E, $peach) Enum>>

package "Document" as Document <<Frame>> {
    Entity(DocumentEntity) {
        id: string
        data: Buffer
        modified: boolean
        revisions: RevisionEntity[]
    }

    Entity(RevisionEntity) {
        id: string
        document: DocumentEntity
        name: string | null
        content: @blocknote/core.Block[]
        createdAt: Date
        deletedAt: Date | null
    }

    Type(AttachmentUploadUrl) {
        uploadUrl: string
        publicUrl: string
    }

    Type(PaginatedRevisions) {
        revisions: RevisionEntity[]
        page: number
        limit: number
        total: number
    }

    Type(OutgoingLinksAndTags) {
        tags: string[]
        outgoingLinkIds: string[]
    }

    Service(DocumentService) {
        - toYDoc(entity: DocumentEntity): yjs.Doc
        - bufferToBlockNote(data: Buffer, editor: ServerBlockNoteEditor): @blocknote/core.Block[]
        - extractTagsAndOutgoingLinkIds(editor: ServerBlockNoteEditor): OutgoingLinksAndTags
        + commitDocument(documentId: string, userId: string)
        + getAttachmentUploadUrl(documentId: string, userId: string): AttachmentUploadUrl
        + updateDataById(id: string, data: Buffer)
        + getById(id: string): DocumentEntity | null
    }

    Service(RevisionService) {
        + getRevision(revisionId: string): RevisionEntity
        + getRevisionsByDocumentId(documentId: string, page: number, limit: number): PaginatedRevisions[]
        + renameRevision(revisionId: string, name: string | null)
        + deleteRevision(revisionId: string)
    }

    Enum(UserNotePermission) {
        read
        write
        delete
    }

    Type(WorkspaceItemPermission) {
        canRead: boolean
        canWrite: boolean
        canDelete: boolean
    }

    Service(AuthorizationService) {
        - toWorkspaceItemPermissionPb(permissions: UserNotePermission): pb.WorkspaceItemPermission
        + hasNotePermission(memberId: string, documentId: string, permission: UserNotePermission): boolean
        + getWorkspaceItemPermission(memberId: string, workspaceId: string): WorkspaceItemPermission
        + getUserNotePermissions(memberId: string, documentId: string): WorkspaceItemPermission
    }

    Type(NoteModel) {
        id: string
        name: string
        icon?: string
        folderId: string
        tags: string[]
        updatedAt?: Date
        trashed?: TrashedModel
    }

    Type(TrashedModel) {
        by: 'purpose' | 'parent'
        at: Date
    }

    Type(WorkspaceModel) {
        id: string
        name: string
        slug: string
    }

    Service(NoteService) {
        + getNoteById(noteId: string, userId: string, excludeTrashed?: boolean): NoteModel
        + getNoteName(noteId: string): string
        + getWorkspaceByNote(userId: string, noteId: string): WorkspaceModel
        - toTrashedModel(trashed: Trashed): TrashedModel
    }

    Service(HocuspocusService) {
        + onRoleChanged(workspaceId: string, userId: string)
        + onMemberRemoved(workspaceId: string, userId: string)
    }

    Service(StorageService) {
        + generateAttachmentPresignedUploadUrl(key: string): AttachmentUploadUrl
    }
}
@enduml
```

<!-- diagram id="class-diagram-document" -->

<!-- vim:set tabstop=4 softtabstop=4 shiftwidth=4: -->
