---
title: Document
order: 2
---

# Document Service

NestJS service holding the collaborative (Yjs / Blocknote) content of notes and serving the
realtime websocket. Note content is committed back to the Note service through events.

- [Domain](./domain): class diagrams
- [Data](./data): database schema
- [Flows](./flows): collaboration, commits, and deletion

## Use cases

| Use case                        | Behavior                                                                                                                                                                                                        |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Collaborate on a note           | Connect using a JWT and note read permission; users without write permission and users viewing trashed notes connect read-only. Edits are synchronized as Yjs updates and stored with a debounce.               |
| Commit content                  | A writer can commit manually. Eligible plans also commit after stores or sweep idle modified documents. A commit creates a revision and publishes content, tags, and outgoing links for downstream consumers.   |
| Manage revision history         | List revisions with pagination and an optional trigger filter, read a snapshot with content, rename it, or hard-delete it. A retention job purges old revisions according to the plan while keeping the latest. |
| Upload an attachment            | A writer requests a presigned upload URL and public URL for an attachment associated with the document.                                                                                                         |
| React to access or note changes | Workspace role and membership events update or close live connections. A note deletion closes connections and removes its document and revisions.                                                               |

HTTP authentication trusts gateway-forwarded user headers; the WebSocket validates its own JWT.
The revision handlers currently do not perform a per-document authorization check; they should not
be assumed to enforce the same permissions as commit or attachment upload.

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

modules: Modules {
  api: Api (OpenAPI Generator)
  authorization: Authorization
  blocknote: Blocknote
  database: Database
  document: Document
  hocuspocus: Hocuspocus
  note: Note
  revision: Revision
  storage: Storage

  api -> document
  api -> revision
  document -> blocknote
  document -> storage
  revision -> blocknote
  revision -> authorization
  hocuspocus -> authorization
  hocuspocus -> note
}
```

<!-- diagram id="architecture-document-diagram" -->
