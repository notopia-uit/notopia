# Document Service

## Service Purpose

The Document service manages document content for Notopia notes. It is responsible for:

- **Document Editing** — Storing and updating structured rich text content using the BlockNote editor format
- **Real-Time Collaboration** — Yjs CRDT-based collaborative editing via WebSocket with awareness of note permissions and membership changes
- **Revision History** — Creating snapshots of document state on commit, with paginated reading, renaming, and soft-deleting of revisions
- **Attachment Storage** — Generating presigned URLs for document attachments
- **Integration Events** — Emitting events on document commit containing metadata for downstream services
- **Authorization Enforcement** — Consuming workspace role and membership change events to enforce permissions on active collaborative sessions

## Architecture Pattern

Modular NestJS monolith with microservice hybrid (HTTP REST + Kafka consumer/producer + WebSocket).

### Module Responsibilities

| Module | Responsibility |
|---|---|
| **Document** | Core document entity, commit workflow, attachment URL generation, event publishing |
| **Revision** | Revision history CRUD (create on commit, rename, soft-delete, paginated read) |
| **Hocuspocus** | Yjs collaborative editing via WebSocket, connection auth/permission enforcement, event consumption |
| **BlockNote** | BlockNote schema provisioning, YDoc↔blocks conversion utilities |
| **Note** | gRPC client to Note service for note metadata and workspace lookups |
| **Authorization** | gRPC client to Authorization service for permission checks |
| **Authentication** | Multi-tenant JWKS token validation |
| **Storage** | S3-compatible presigned URL generation for attachments |
| **Kafka** | Global Kafka client (producer + consumer) |
| **Database** | TypeORM DataSource configuration (PostgreSQL) |
| **Config** | Zod-validated environment configuration |
| **Common** | HTTP user guard, request user decorator, global exception filter |

## High-Level Structure

```
apps/document/
├── src/
│   ├── Bootstrap                  — NestJS application entry and root module
│   ├── document/                  — Core document CRUD + commit + event emit
│   ├── revision/                  — Revision history with pagination + soft-delete
│   ├── hocuspocus/                — Yjs real-time collaboration (WebSocket)
│   ├── blocknote/                 — BlockNote schema + content conversion
│   ├── note/                       — gRPC client → Note service
│   ├── authorization/             — gRPC client → Authorization service
│   ├── authentication/            — JWKS-based JWT validation
│   ├── storage/                   — S3 presigned URL generation
│   ├── kafka/                     — Global Kafka client
│   ├── database/                  — TypeORM DataSource
│   ├── config/                    — Zod config schemas
│   └── common/                    — Guards, decorators, exception filters
├── database/                      — Migrations + seed data
├── api/document/                  — OpenAPI spec
└── Build & Deploy Config          — Docker and Nx configuration
```

## API Contracts Served

The service exposes two primary API contracts:

- **HTTP REST** via OpenAPI specification (located in `api/document/`), generated into NestJS server code. Provides document and revision CRUD operations.
- **WebSocket** for real-time collaborative editing using the Yjs sync protocol. Clients authenticate via JWT token and receive read/write permission enforcement based on workspace roles.

## Communication Pattern

| Channel | Direction | Purpose |
|---|---|---|
| **HTTP REST** | External → Service | Document commit, attachment upload URL, revision CRUD |
| **WebSocket (Yjs)** | Browser ↔ Service | Real-time collaborative editing with CRDT sync |
| **Kafka (Producer)** | Service → Other Services | Document commit events with metadata for indexing and linking |
| **Kafka (Consumer)** | Authorization Service → Service | Role change and membership removal events for permission enforcement |
| **gRPC (Client)** | Service → Note Service | Note metadata and workspace lookups |
| **gRPC (Client)** | Service → Authorization Service | Permission checks for read/write access |
| **S3** | Service → Storage | Presigned upload URLs for attachments |

## Key Technologies & Patterns

### Modular NestJS Architecture
Feature modules (document, revision, hocuspocus, blocknote, note, authorization, authentication, storage, kafka, database) each with their own providers and exports. Root module imports all feature modules and registers the generated API server with concrete implementations.

### TypeORM with PostgreSQL
Document and revision entities with pessimistic locking on commit to prevent concurrent modifications. Soft-delete support for revisions. One-to-many relationship between documents and revisions with cascade delete.

### Hocuspocus Yjs Server
Real-time collaborative editing using Yjs CRDT with Hocuspocus server. Document state persisted to PostgreSQL. Authentication hook validates JWT and checks permissions. Connection management handles role changes and membership removals by updating read-only mode or closing connections.

### BlockNote Editor Format
Structured rich text content stored as BlockNote blocks. Conversion utilities between Yjs binary format and BlockNote JSON. Tag and link extraction for integration events.

### S3-Compatible Storage
Presigned URL generation for time-limited document attachment uploads. Supports local and cloud S3-compatible backends.

### Kafka for Event-Driven Communication
Global Kafka client for both producing document commit events and consuming authorization events. Consumer group-based consumption with retry logic.

### Generated API Server from OpenAPI
Abstract base classes generated from OpenAPI specification. Concrete implementations provide business logic.

### Generated gRPC Clients from Proto
Type-safe gRPC clients for Note and Authorization services generated from protobuf definitions.

### Multi-Tenant JWKS Authentication
JWT token validation supporting multiple JWKS endpoints and issuers. User identity extracted from token claims.

### HTTP User Extraction via Forwarded Headers
User identity extracted from forwarded headers set by API gateway. Applied globally via HTTP guard.

## Dependencies

| Dependency | Purpose |
|---|---|
| **PostgreSQL** | Document and revision data storage |
| **Kafka** | Integration event bus (producer + consumer) |
| **S3-Compatible Storage** | Document attachment file storage |
| **Note Service** | gRPC-based note metadata lookup |
| **Authorization Service** | gRPC-based permission checks |
| **JWKS Identity Provider** | Multi-tenant JWT token validation |
