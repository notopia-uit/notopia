# Search Worker

## Service Purpose

The Search Worker indexes note and document data into Meilisearch to power full-text search across the Notopia application. It is responsible for:

- **Note Indexing** — Indexing note metadata (name, workspace, folder, trashed state) on creation and update
- **Document Content Indexing** — Converting BlockNote structured content to searchable text on document commit
- **Search Index Maintenance** — Maintaining a searchable index that the frontend queries via scoped tenant tokens

## Architecture Pattern

Simple NestJS Kafka consumer (microservice pattern). The service has no REST API beyond the implicit health endpoint. It is a pure event-driven consumer.

### Component Structure

```
AppModule (root)
├── ConfigModule       — Zod-validated env config
├── LoggerModule       — Structured logging
├── OpenTelemetryModule — Observability
├── Providers
│   ├── Meilisearch    — Search index client
│   ├── BlockNote Schema — Content parsing
│   └── AppService     — Indexing business logic
└── Controllers
    └── AppController  — Kafka event handlers
```

## High-Level Structure

```
apps/search-worker/
├── src/
│   ├── Bootstrap                  — NestJS application entry and root module
│   ├── Configuration              — Zod schemas and environment validation
│   ├── Kafka Transport            — Kafka microservice configuration
│   ├── Event Handlers             — Kafka event consumer logic
│   ├── Indexing Service           — Search index operations
│   ├── Observability              — OpenTelemetry setup
│   └── Error Handling             — Retry logic and exception filters
├── seed/                          — Seed data scripts
├── Build & Deploy Config          — Docker and Nx configuration
└── Dependencies                   — Package manifest
```

## API Contracts Served

None. The Search Worker exposes no HTTP API beyond the default health endpoint. It is a pure Kafka consumer.

### Kafka Topics Consumed

The service listens to three integration event topics:

- Note creation events — Index new note metadata
- Note update events — Update indexed note metadata
- Document commit events — Index document content

## Communication Pattern

| Channel | Direction | Purpose |
|---|---|---|
| **Kafka (Consumer)** | Note Service / Document Service → Worker | Listens to integration events for note lifecycle and document commits |
| **Meilisearch** | Worker → Search Index | Indexes note metadata and converted document content |
| **Health Endpoint** | External → Worker | Default NestJS health check |

## Key Technologies & Patterns

### NestJS Kafka Microservice
Event-driven consumer using NestJS microservices pattern. Event handlers receive typed payloads from generated API types. Error handling with retry logic and max-retry exception filter.

### BlockNote Content Conversion
BlockNote schema for parsing structured content. Conversion to searchable text format (Markdown) for full-text indexing without storing full JSON structure.

### Meilisearch for Full-Text Search
Single search index updated via upsert semantics. Supports scoped tenant tokens for frontend queries.

### Generated Event Types
Event payload types imported from shared API generation package, ensuring type safety across service boundaries.

### Zod Configuration Validation
All configuration (app, kafka, meilisearch) defined as Zod schemas and validated at startup.

### OpenTelemetry Observability
Auto-instrumentation for Kafka, logging, and runtime metrics.

### Structured Logging
JSON logging with pretty-print in development. Exception serialization includes stack traces.

## Dependencies

| Dependency | Purpose |
|---|---|
| **Kafka** | Integration event consumer |
| **Meilisearch** | Full-text search index |
