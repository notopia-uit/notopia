# Note

:::info

- `Note` only store the info
- `Document` is the `Note` content in binary format for editing and collaboration

:::

## Create Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant AS as Authorization Service
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service
    actor SU as Subcribed Users

    U->>+NS: CreateNote
    NS->>+AS: HasWorkspacePermission
    break Authorization Failed
        AS-->>NS: No Permission
        NS-->>U: No Permission
    end
    AS-->>-NS: Ok
    NS->>NS: Create Note
    par Response
        NS-->>-U: Note ID
    and NoteCreated domain event
        loop Domain NoteCreatedEvent not published
            NS-)MB: Publish domain NoteCreatedEvent
        end
        MB-)+NS: Domain NoteCreatedEvent
        par Publish workspace event
            loop Not processed domain NoteCreatedEvent
                NS->>NS: Convert to workspace NoteCreatedEvent
            end
            NS-)SU: Publish workspace NoteCreatedEvent
        and Publish NoteCreatedEvent integration event
            loop Not processed domain NoteCreatedEvent
                NS->>NS: Convert to NoteCreatedEvent
            end
            NS-)-MB: Publish integration NoteCreatedEvent
            MB-)+SW: Retrieve integration NoteCreatedEvent
            SW->>SW: Process NoteCreatedEvent
            SW->>+SS: Index Note
            par Ack
                SS-->>-SW: Ok
            and Indexing Note
                SS->>+SS: Index Note
            end
        end
    end
```

## Get Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant DS as Document Service
    participant AS as Authorization Service

    U->>+NS: GetNote
    NS->>NS: Get Note
    NS->>+AS: HasNotePermission
    break Authorization Failed
        AS-->>NS: No Permission
        NS-->>U: No Permission
    end
    AS-->>-NS: Ok
    NS-->>-U: Note
    opt Enter Document
        U->>+DS: WsDocument
        DS->>DS: Get Document
        opt Document does not exists
            DS->>+NS: CheckNoteExistence
            NS->>NS: Check Note existence
            alt Note does not exists
                NS-->>-DS: Not found
                DS-->>U: Not found
            else Note exists
                DS->>DS: Init document
            end
        end
        DS-->>U: Hocuspocus Connection
        loop Edit debounce by handled by Hocuspocus
            U->>DS: Edit Document
            DS->>DS: Save Document changes, broadcast to other clients
        end
    end
```

## Commit Document

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant DS as Document Service
    participant MB as Message Broker
    participant NS as Note Service
    participant SW as Search Worker
    participant SS as Search Service

    U->>+DS: CommitDocument
    DS->>+DS: Create Document revision
    par Response
        DS-->>-U: Ok
    and Publish integration DocumentCommittedEvent
        DS-)MB: Publish integration DocumentCommittedEvent
    end
    par Note service update note
        loop Not processed integration DocumentCommittedEvent
            MB-)NS: Receive DocumentCommittedEvent
        end
        NS->>NS: Update note size, tags
    and Search service update note index
        loop Not processed integration DocumentCommittedEvent
            MB-)SW: Receive DocumentCommittedEvent
        end
        SW->>SW: Convert to markdownContent, tags in form of NoteSearch
        SW->>+SS: Index Note
        par Ack
            SS-->>-SW: Ok
        and Indexing Note
            SS->>+SS: Index Note
        end
    end
```

## Update Note

:::info

- This also include the trash and restore note

:::

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant AS as Authorization Service
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: UpdateNote
    NS->>+AS: HasNotePermission
    break Authorization Failed
        AS-->>NS: No Permission
        NS-->>U: No Permission
    end
    AS-->>-NS: Ok
    NS->>NS: Update Note
    par Response
        NS-->>-U: Ok
    and Publish integration NoteUpdatedEvent
        loop Not published integration NoteUpdatedEvent
            NS-)MB: Publish NoteUpdatedEvent
        end
        loop Not processed integration NoteUpdatedEvent
            MB-)SW: NoteUpdatedEvent
        end
        SW->>+SS: Index Note
        par Ack
            SS-->>-SW: Ok
        and Indexing Note
            SS->>SS: Index Note
        end
    end
```

## Permanently Delete Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant AS as Authorization Service
    participant MB as Message Broker
    participant DS as Document Service
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: PermanentlyDeleteWorkspaceItems (note)
    NS->>+AS: HasWorkspacePermission
    break Authorization Failed
        AS-->>NS: No Permission
        NS-->>U: No Permission
    end
    NS->>NS: Permanently delete Note
    par Response
        NS-->>-U: Ok
    and Publish NoteDeletedEvent
        NS->>MB: Publish NoteDeletedEvent
    end
    par Process NoteDeletedEvent
        MB->>DS: NoteDeletedEvent
        DS->>DS: Delete Document and Revisions
    and Search
        MB->>SW: NoteDeletedEvent
        SW->>+SS: Delete Note from index
        SS->>+SS: Delete Note
    end
```

<!-- vim:set tabstop=4 shiftwidth=4: -->
