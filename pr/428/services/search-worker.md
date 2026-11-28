---
url: /notopia/pr/428/services/search-worker.md
---

# Search Worker

NestJS worker consuming integration events from Redpanda and syncing notes into Meilisearch.

* [Flows](./flows): indexing, deletion races, and readiness.
* [Note service](../note/) owns note metadata and search-token authorization;
  [Document service](../document/) owns editable content.

## Use cases

| Event topic                                      | Index effect                                                            |
| ------------------------------------------------ | ----------------------------------------------------------------------- |
| `events.integration.note.note.created`           | Upsert note ID, name, and workspace ID.                                 |
| `events.integration.note.note.updated`           | Upsert name, folder ID and name, and trashed state.                     |
| `events.integration.document.document.committed` | Convert BlockNote blocks to lossy Markdown and upsert content and tags. |
| `events.integration.note.note.deleted`           | Mark a tombstone and delete the note from the index.                    |

Updates are partial and can arrive independently: metadata and committed content populate the
same `notes` document by ID. Indexing is asynchronous; uncommitted edits are not indexed.
The worker does not serve search requests. The note service checks workspace read permission
before issuing a scoped Meilisearch tenant token.

## Operations

* Run `nx run search-worker:seed` in every environment to configure the `notes` index (primary
  key `id`), searchable fields (`name`, `folderName`, `plainTextContent`, `tags`), filters
  (`workspaceId`, `tags`), name sorting, and a search-only key. Set
  `NOTOPIA_SEARCH_WORKER_MEILI_HOST`, `NOTOPIA_SEARCH_WORKER_MEILI_API_KEY`, and
  `NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID`; the UID must match
  `NOTOPIA_NOTE_MEILISEARCH_NOTE_SEARCH_KEY_UID`. A newly created key is printed only outside
  production.
* Development fixtures are separate from seed: run `nx run search-worker:fixtures:transform`
  before `nx run search-worker:fixtures`.
* HTTP health endpoints `/search-worker/health/live` and `/search-worker/health/ready` listen
  on `NOTOPIA_SEARCH_WORKER_HEALTH_CHECK_PORT` (default `28083`). Readiness requires both
  Meilisearch health and a recent Kafka group join or heartbeat.
* Set `NOTOPIA_SEARCH_WORKER_REDIS_URL` for shared deletion tombstones across instances.
  Without it, the worker warns and uses a single-instance in-memory LRU cache. Tombstones
  expire after `NOTOPIA_SEARCH_WORKER_TOMBSTONE_TTL_SECONDS` (default `86400`); choose a TTL
  longer than worst-case consumer lag plus retries. `NOTOPIA_SEARCH_WORKER_TOMBSTONE_MAX`
  bounds the in-memory cache, so eviction can also end protection early.

Meilisearch write tasks are awaited and unsuccessful tasks are reported as errors. The source
contains a retry exception filter, but it is not registered by the worker; do not assume its
retry or offset policy is active.

## Architecture

```d2
direction: right

vars: {
  d2-config: {
    layout-engine: elk
    theme-overrides: {
      N1: "#4c4f69"
      N2: "#5c5f77"
      N4: "#acb0be"
      N5: "#ccd0da"
      N7: "#eff1f5"
      B1: "#4c4f69"
      B2: "#6c6f85"
      B3: "#bcc0cc"
      B4: "#ccd0da"
      B5: "#dce0e8"
      B6: "#eff1f5"
      AA4: "#1e66f5"
      AA5: "#7287fd"
      AB4: "#8839ef"
      AB5: "#dc8a78"
    }
  }
}

controller: Controller (Kafka)
service: Service
meili: MeiliSearch

controller -> service
service -> meili
```
