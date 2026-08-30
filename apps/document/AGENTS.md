# Document Service (nx: `document`)

Owns document *content* (note metadata lives in `internal/note`). NestJS: HTTP REST +
Yjs WebSocket + Kafka producer/consumer. Spec: `api/document/`.

## Modules (`src/`)

| Module | Note |
| --- | --- |
| `document/` | `DocumentEntity` (Yjs state as `bytea` + `modified` flag), commit workflow, attachment URLs |
| `revision/` | Revision snapshots, paginated read, rename, soft-delete |
| `hocuspocus/` | Yjs server, connection auth, live permission revocation |
| `blocknote/` | Schema provider + YDoc↔blocks conversion, tag/link extraction |
| `note/`, `authorization/` | gRPC clients (generated from `proto/`) |
| `authentication/` | Multi-issuer JWKS validation — for the **WebSocket** token |
| `common/` | `UserGuard` + `@User()` — for **HTTP**, reads gateway-forwarded headers |
| `storage/`, `kafka/`, `database/`, `config/` | S3 presign, Kafka client, TypeORM, Zod config |

Two auth paths, easy to trip on: HTTP trusts Traefik's forwarded headers; the WebSocket
gets a raw JWT in the connect payload and validates it itself in
`Hocuspocus.onAuthenticate`.

## The commit pipeline

`DocumentService.commitDocument()` is the hinge of the whole system. In one transaction
(pessimistic write lock on the document) it: saves a `RevisionEntity` of the current
blocks, clears `modified`, and emits
`events.integration.document.document.committed`.

Downstream: `internal/note` (`e_documentcommitted.go`) updates note size/tags/outgoing
links; `search-worker` indexes the content into Meilisearch. **Nothing else re-indexes** —
if a document is never committed, its content is never searchable and its links never
appear in the graph.

Commit is currently **manual only** (`POST /document/documents/{id}/commit`).

## Yjs / Hocuspocus

- `HocuspocusDatabase.store` persists Yjs state and sets `modified = true`; there is no
  debounce configured, so it fires on the extension's default cadence.
- The `metadata` Y.Map (`YDocMetadataMap`) carries `{ modified }` to clients;
  `HocuspocusService.setModified` is how the server pushes that state after a commit.
- `onRoleChanged` / `onMemberRemoved` consume Kafka events from `authorization` and walk
  live connections to flip `readOnly` or close with codes 4001/4002.

## Conventions

- HTTP handlers extend the generated abstract class from
  `@notopia-uit/api-document-nestjs-server` — change `api/document/` and run `nx gen api`
  first, then implement the new method.
- Kafka payload types come from `@notopia-uit/api-share-gen` (defined in `api/share/`),
  never hand-written interfaces.
- `@Traceable()` on services; structured logger calls are `this.logger.log({...}, 'msg')`.
