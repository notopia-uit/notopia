---
title: Note
order: 1
---

# Note Service

Go service owning workspaces, folders, note metadata, links, tags and trash. DDD + CQRS,
with a transactional outbox for integration events. Note only stores note metadata; the
editable binary content lives in the [Document service](../document/).

- [Domain](./domain): class diagrams
- [Data](./data): database schema
- [Flows](./flows): sequence diagrams
- [Events](./events): topics published and consumed

## Use cases

| Area              | Operations and behavior                                                                                                      |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Workspaces        | Create, rename, change slug, delete, list and look up workspaces; creation also creates a root folder and owner membership.  |
| Membership        | List and update members, leave a workspace and search users. Member limits and payer protection apply to membership changes. |
| Folders and notes | Create, rename, change icons and move items; read notes and the workspace tree. Metadata is separate from document content.  |
| Trash             | Trash, restore, permanently delete or empty trash; a scheduled retention purge removes expired trash.                        |
| Discovery         | Read links and graphs, and request a workspace-scoped search token. Graphs are subject to billing entitlements.              |
| Document commits  | Consume committed content to update note size, tags and outgoing links.                                                      |

See [Flows](./flows) for authorization, billing, deletion and asynchronous event paths.
Note and workspace publishing handlers are currently stubs, not implemented publishing workflows.

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

controllers: Controllers {
  http: HTTP / REST Controller
  grpc: gRPC Server
  event: Event Processor
}

application: Application {
  command_handler: Command Handler
  query_handler: Query Handler
  event_handler: Event Handler
  authorization_svc_interface: Authorization Svc Interface
  identity_svc_interface: Identity Svc Interface
  workspace_event_bus_interface: Workspace Event Bus Interface
  workspace_events: Workspace Events
  integration_event_publisher_interface: Integration Event Publisher Interface
  integration_events: Integration Events
}

domain: Domain {
  model: Domain Model
  repository_interface: Repository Interface
  domain_event: Domain Events
}

infrastructure: Infrastructure {
  persistence: Persistence/SQL
  authentik_svc: Authentication Svc
  authorization_svc: Authorization Svc
  outbox: Outbox
  workspace_event_bus: Workspace Event Bus
  integration_event_publisher: Integration Event Publisher
}

application -> domain
controllers -> application
infrastructure -> application
infrastructure -> domain
```

<!-- diagram id="architecture-note-diagram" -->
