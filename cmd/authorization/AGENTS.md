# Authorization Service

## 1. Service Purpose

The Authorization service provides centralized access control for the Notopia workspace system. It is responsible for:

- **Workspace membership** — Managing user roles within workspaces (owner, editor, viewer)
- **Permission enforcement** — Answering authorization questions for workspaces and workspace items (notes, folders)
- **Role-based access control** — Defining and enforcing role hierarchies and resource permissions
- **Resource inheritance** — Ensuring workspace items inherit permissions from their parent workspace
- **Integration events** — Publishing membership changes to downstream consumers

## 2. Architecture Pattern

Clean Architecture + CQRS, adapted for policy-based authorization. Unlike traditional Domain-Driven Design with aggregates, the authorization model is driven by a policy engine (Casbin) that manages role-based access control rules.

| Layer | Package | Responsibility |
|-------|---------|-----------------|
| **Application** | `internal/authorization/app/` | CQRS command/query handlers wrapping policy engine operations, policy bootstrapping, integration event types |
| **Controller** | `internal/authorization/controller/` | Adapters: gRPC service handlers, health check endpoints |
| **Infrastructure** | `internal/authorization/infra/` | Policy engine adapter, database connection, event publishing |
| **Config** | `internal/authorization/config/` | Environment-based configuration loading |
| **Errors** | `internal/authorization/errs/` | Domain-specific error types mapped to gRPC status codes |

## 3. High-Level Structure

```
cmd/authorization/          — Service entrypoint, dependency injection setup, Docker configuration
internal/authorization/
├── app/                     — CQRS handlers, policy engine setup, policy definitions, integration event types
├── controller/              — gRPC service handlers and health check endpoints
├── infra/                   — Policy engine adapter, database connection, event publishing
├── config/                  — Configuration loading
└── errs/                    — Error type definitions
```

## 4. API Contracts Served

The service exposes a single protocol:

- **gRPC** — Inter-service communication only. Other services call this service to check permissions, query workspace membership, and manage roles. All requests are validated and instrumented with OpenTelemetry tracing.

Health checks are exposed via HTTP for orchestration and monitoring purposes.

## 5. Communication Pattern

| Channel | Direction | Purpose |
|---------|-----------|---------|
| **gRPC** | Other Services → Service | Permission checks, membership queries, role management |
| **Kafka (Integration Events)** | Service → Other Services | Workspace membership changes for downstream processing |

## 6. Key Technologies & Patterns

**Policy-Based Authorization** — A policy engine (Casbin) manages role-based access control rules. Policies define which roles can perform which actions on which resource types. This approach is more flexible than hardcoded permission logic.

**Resource Inheritance** — Workspace items (notes, folders) inherit permissions from the workspace resource type. Permission checks on items automatically delegate to workspace-level policies.

**Transactional Policy Updates** — Policy mutations are atomic. Multi-step operations (e.g., updating multiple role assignments) are wrapped in transactions to ensure consistency.

**Batch Permission Checks** — Multiple permission checks can be performed in a single operation for efficiency.

**Policy Bootstrap** — On startup, base permission policies and resource inheritance rules are loaded into the policy engine. Development environments can optionally seed additional test data.

**CQRS Pattern** — Commands mutate policies and return errors. Queries read policies and return results. All handlers are decorated with structured logging and distributed tracing.

**Dependency Injection via Wire** — Configuration, policy engine, and handlers are wired together at startup.

## 7. Dependencies

| Dependency | Purpose |
|------------|---------|
| **PostgreSQL** | Policy storage via policy engine adapter |
| **Kafka** | Event publishing for membership changes |
| **Policy Engine** | Role-based access control rule management and enforcement |
