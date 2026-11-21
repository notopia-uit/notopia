---
title: Search Worker Flows
order: 2
---

# Search Worker Flows

These flows describe the asynchronous projection into Meilisearch, not a transactional search
index. See the [worker overview](./) for event topics, configuration, and operating limits.

## Index metadata and committed content

```mermaid
sequenceDiagram
    participant P as Note or Document service
    participant K as Redpanda
    participant W as Search worker
    participant T as Tombstone store
    participant M as Meilisearch notes

    P->>K: Publish created, updated, or committed event
    K->>W: Deliver event
    opt Document committed
        W->>W: Convert BlockNote blocks to lossy Markdown
    end
    W->>T: Check note ID
    alt Tombstone exists
        W->>W: Skip upsert
    else No tombstone
        W->>M: Upsert partial document by ID
        W->>M: Wait for indexing task
        alt Task fails
            W-->>K: Handler throws error
        else Task succeeds
            W->>T: Check note ID again
            opt Deleted during upsert
                W->>M: Delete document and wait for task
            end
        end
    end
```

Created events supply `name` and `workspaceId`; updates supply name, folder, and trashed
state; commits supply tags and `plainTextContent`. These partial writes can arrive in either
order. Search can lag events, and edits are not reflected until a document commit.

## Delete and late events

```mermaid
sequenceDiagram
    participant K as Redpanda
    participant D as Delete handler
    participant U as Concurrent or late upsert
    participant T as Tombstone store
    participant M as Meilisearch notes

    K->>D: note.deleted
    D->>T: Mark ID with TTL
    D->>M: Delete document
    D->>M: Wait for delete task
    K->>U: created, updated, or committed event
    U->>T: Check ID before write
    alt Tombstone present
        U->>U: Skip write
    else Deletion races with write
        U->>M: Upsert and wait for task
        U->>T: Recheck ID
        opt Tombstone now present
            U->>M: Delete document and wait for task
        end
    end
```

The second check handles a deletion that happens after the first check. Protection is bounded
by tombstone lifetime and, for memory storage, cache capacity and process lifetime; Redis
shares tombstones between worker instances. Events arriving after protection expires can
recreate a deleted record.

## Health and recovery

```mermaid
flowchart TD
    A[GET /search-worker/health/ready] --> B{Recent Kafka group join or heartbeat?}
    A --> C{Meilisearch health succeeds?}
    B -->|Yes, less than 15 seconds old| D{Both healthy?}
    B -->|No| D
    C --> D
    D -->|Yes| E[Ready]
    D -->|No| F[Not ready]
```

Disconnect, stop, crash, rebalance, and request timeout mark Kafka down. A subsequent group
join or heartbeat restores the Kafka indicator. The HTTP listener starts independently of
Kafka, and Kafka startup failures are retried. Liveness reports process health without checking
these dependencies.
