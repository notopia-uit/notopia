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
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: CreateNote
    NS->>NS: Create Note
    par Response
        NS-->>-U: Note ID
    and NoteCreated event
        NS-)MB: Publish NoteCreatedEvent
        MB-)SW: Retrieve NoteCreatedEvent
        SW->>SW: Process NoteCreatedEvent
        SW->>+SS: Index Note info
        SS->>+SS: Index Note
    end
```

## Get Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant DS as Document Service

    U->>+NS: GetNote
    NS->>NS: Get Note
    NS-->>-U: Note
    opt Enter Edit Mode
        U->>+DS: WsDocument
        DS->>DS: Establish Hocuspocus connection
        alt Document exists
            DS-->>U: Connection
        else Document does not exist
            DS->>+NS: CheckNoteExistence
            NS->>NS: Check Note existence
            alt Note exists
                DS->>DS: Init document
                DS-->>U: Document binary
            else Note does not exist
                NS-->>-DS: Not found
                DS-->>U: Not found
            end
        end
        loop Edit Document
            U->>DS: Edit Document
            DS->>DS: Save Document changes, broadcast to other clients
        end
    end
```

## Commit Note

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
    DS->>DS: Create Document revision
    par Response
        DS-->>-U: Ok
    and Publish DocumentCommittedEvent
        DS->>MB: Publish DocumentCommittedEvent
    end
    par Process DocumentCommittedEvent
        MB->>SW: DocumentCommittedEvent
        SW->>SW: Update note size, tags
    and
        MB->>SW: NoteUpdated event
        SW->>SW: Convert to markdownContent in form of NoteSearch
        SW->>+SS: Update Note size, tags
        SS->>SS: Update Note index
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
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: UpdateNote
    NS->>NS: Update Note
    par Response
        NS-->>-U: Ok
    and Publish NoteUpdatedEvent
        NS->>MB: Publish NoteUpdatedEvent
        MB->>SW: NoteUpdatedEvent
        SW->>+SS: Update Note info
        SS->>+SS: Update Note index
    end
```

## Permanently Delete Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant MB as Message Broker
    participant DS as Document Service
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: PermanentlyDeleteWorkspaceItems (note)
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
