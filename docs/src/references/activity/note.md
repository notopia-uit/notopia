# Note

:::info

- `Note` includes the tree structure content _(BlockNoteJS model)_, which is used for feeding data into the editing action
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

    U->>+NS: Create Note
    NS->>NS: Create Note
    par Response
        NS-->>-U: ID of created Note
    and Create NoteCreated event
        NS->>MB: Publish NoteCreated event
        MB->>SW: NoteCreated event
        SW->>SW: Process NoteCreated event
        SW->>+SS: Index Note
        SS--)-NS: Acknowledgement
        SS->>+SS: Index Note
    end
```

## Import Document

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant DS as Document Service
    participant NS as Note Service
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service
```

## Get Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant DS as Document Service

    U->>+NS: Get Note content
    NS->>NS: Get Note content
    NS-->>-U: Note content
    opt Enter Edit Mode
        U->>+DS: Get Document for editing
        DS->>DS: Get Document
        alt Document exists
            DS-->>U: Document binary
        else Document does not exist
            DS->>+NS: Get Note content
            NS->>NS: Get Note content
            alt Note exists
                NS-->>DS: Note content
                DS->>DS: Create Document
                DS-->>U: Document binary
            else Note does not exist
                NS-->>-DS: Not found
                DS-->>-U: Not found
            end
        end
    end
```

## Commit Note

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant DS as Document Service
    participant NS as Note Service
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service

    U->>+DS: Commit Document changes
    DS->>DS: Process Document changes
    DS->>+NS: Apply Document changes to Note
    NS->>NS: Update Note content, create new revision
    par Response
        NS-->>-DS: Ok
        DS-->>-U: Ok
    and Publish NoteUpdated event
        NS->>MB: Publish NoteUpdated event
        MB->>SW: NoteUpdated event
        SW->>SW: Process NoteUpdated event
        SW->>+SS: Update Note index
        SS--)-NS: Acknowledgement
        SS->>+SS: Update Note index
    end
```

## Update Note

:::info

Update Note mean just update the Note metadata, like name, folderId (move), trashed, icon...

:::

```mermaid
sequenceDiagram
    autonumber

    actor U as User
    participant NS as Note Service
    participant MB as Message Broker
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: Update Note metadata
    NS->>NS: Update Note metadata
    par Response
        NS-->>-U: Ok
    and Publish NoteUpdated event
        NS->>MB: Publish NoteUpdated event
        MB->>SW: NoteUpdated event
        SW->>SW: Process NoteUpdated event
        SW->>+SS: Update Note index
        SS--)-NS: Acknowledgement
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
    participant SW as Search Worker
    participant SS as Search Service

    U->>+NS: Permanently delete Note
    NS->>NS: Permanently delete Note
    par Response
        NS-->>-U: Ok
    and Publish NoteDeleted event
        NS->>MB: Publish NoteDeleted event
        MB->>SW: NoteDeleted event
        SW->>SW: Process NoteDeleted event
        SW->>+SS: Delete Note from index
        SS--)-NS: Acknowledgement
        SS->>+SS: Delete Note from index
    end
```

<!-- vim:set tabstop=4 shiftwidth=4: -->
