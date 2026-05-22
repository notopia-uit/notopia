# Note Service

## 1. Service Purpose

The Note service manages the core content hierarchy and lifecycle of Notopia. It is responsible for:

- **Workspace management** — Creating, renaming, managing membership, publishing/unpublishing for public access
- **Folder hierarchy** — Creating, renaming, moving, and deleting folders within workspace trees
- **Note lifecycle** — Creating, renaming, reordering, tracking relationships (links and backlinks), managing metadata, publishing/unpublishing
- **Real-time notifications** — Broadcasting workspace changes to connected clients via streaming
- **Integration events** — Publishing note lifecycle events to downstream consumers (search indexing, document processing)
- **Search access** — Generating scoped search tokens for frontend access to full-text search

## 2. Architecture Pattern

Clean Architecture + Domain-Driven Design + CQRS + Event-Driven Architecture.

The service separates concerns into distinct layers, each with clear responsibilities:

| Layer | Package | Responsibility |
|-------|---------|-----------------|
| **Domain** | `internal/note/domain/` | Enterprise business rules — aggregates (Workspace, Folder, Note), value objects, domain events, repository interfaces, unit of work pattern |
| **Application** | `internal/note/app/` | CQRS command/query/event handlers, service interfaces (authorization, identity, search), read model definitions, integration event types |
| **Controller** | `internal/note/controller/` | Adapters: HTTP REST handlers, gRPC inter-service handlers, event consumers, health check endpoints |
| **Infrastructure** | `internal/note/infra/` | Implementations: database persistence, transactional outbox, event publishing, real-time event streaming, search integration, identity provider clients, inter-service gRPC clients |
| **Component** | `internal/note/component/` | Shared infrastructure: event bus setup, serialization, validation |
| **Config** | `internal/note/config/` | Environment-based configuration loading and validation |
| **Errors** | `internal/note/errs/` | Domain-specific error types |

## 3. High-Level Structure

```
cmd/note/                    — Service entrypoint, dependency injection setup, Docker configuration
internal/note/
├── domain/                  — Business entities, domain events, repository interfaces, unit of work
├── app/                     — CQRS handlers, service interfaces, read models, integration event types
├── controller/              — HTTP, gRPC, event consumer, and health check adapters
├── infra/                   — Database, event publishing, streaming, search, identity, and inter-service implementations
├── component/               — Shared event bus, serialization, validation
├── config/                  — Configuration loading
└── errs/                    — Error type definitions
```

## 4. API Contracts Served

The service exposes two primary protocols:

- **HTTP REST** — OpenAPI 3.0 specification for external clients (frontend, third-party integrations). Requests are validated against the spec. Authentication is enforced via JWT tokens passed through the API gateway.
- **gRPC** — Inter-service communication for workspace and note metadata queries. Includes request validation and OpenTelemetry instrumentation.

Additionally, the service streams real-time workspace updates via Server-Sent Events (SSE) to connected browsers.

## 5. Communication Pattern

| Channel | Direction | Purpose |
|---------|-----------|---------|
| **HTTP REST** | External → Service | Frontend and third-party API consumption |
| **gRPC** | Service ↔ Authorization Service | Permission checks and workspace member queries |
| **gRPC** | Other Services → Service | Metadata queries from document and search services |
| **Kafka (Integration Events)** | Service → Other Services | Note lifecycle events (created, updated, deleted) for downstream processing |
| **Kafka (Domain Events)** | Service (internal) | Domain events for internal side effects and integration bridging |
| **Redis Pub/Sub** | Service (internal) | Workspace event streaming across service instances |
| **SSE** | Service → Browser | Real-time workspace updates to connected clients |

## 6. Key Technologies & Patterns

**Domain Events on Aggregates** — Each aggregate maintains an internal event queue. Mutating operations record events. After persistence, events are extracted and routed through the event bus for side effects and integration publishing.

**Transactional Outbox** — Domain events are persisted to an outbox table within the same database transaction as the aggregate state. A background process polls this table and publishes events to Kafka, ensuring reliable delivery without distributed transactions.

**CQRS Pattern** — Commands mutate state and return errors. Queries read state and return results. Event handlers process side effects. All handlers are decorated with structured logging and distributed tracing.

**Dependency Injection via Wire** — Two-level setup: internal layer combines domain, application, controller, infrastructure, and configuration; top-level adds service metadata and initializes the server.

**Concurrent Server Components** — HTTP, gRPC, event listeners, real-time event hub, and outbox forwarder run concurrently with coordinated graceful shutdown.

**Real-Time Event Streaming** — Workspace events are published to a distributed stream (Redis), fanned out in-memory per workspace, and pushed to browsers via SSE with heartbeat keepalive.

**Type-Safe Database Queries** — SQL queries are code-generated with full type safety, null handling, and tracing support.

**Handler Decoration** — Generic decorators wrap command/query handlers with structured logging and OpenTelemetry tracing at construction time.

## 7. Dependencies

| Dependency | Purpose |
|------------|---------|
| **PostgreSQL** | Primary data store for workspaces, folders, notes, and transactional outbox |
| **Kafka** | Event bus for integration events and domain event routing |
| **Redis** | Distributed event streaming for real-time workspace updates |
| **Search Service** | Full-text search indexing and scoped token generation |
| **Identity Provider** | User lookup and search |
| **Authorization Service** | Permission checks and workspace member management via gRPC |
| **Document Service** | Consumes integration events for document processing |
| **Search Worker** | Consumes integration events for search index synchronization |
