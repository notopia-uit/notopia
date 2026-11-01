---
title: Messaging
order: 2
---

# Messaging

Services communicate asynchronously through Redpanda (Kafka API).

## Transactional outbox

Go services publish through a Watermill transactional outbox so that a domain write and the
event it produces commit together. The flow in the Note service:

1. The repository `Save` drains domain events and publishes them with the Watermill
   `forwarder` publisher, which writes the message into an outbox table in the same Postgres
   transaction as the aggregate change.
2. The outbox table (default name `eventsToForward`, config `outbox_table_name`) follows the
   Watermill SQL schema adapter and is created at startup (`InitializeSchema`).
3. A Watermill SQL subscriber polls the table (every second) and the forwarder re-publishes each
   message to its real topic on Redpanda.

Because of this, publishing an integration event inside a unit of work is safe. Notifications
that must not fail the request (for example billing's `subscription.changed`) are instead
published best-effort after commit. NestJS services publish and consume directly with
`@nestjs/microservices`.

## Topic naming

| Kind                          | Pattern                                            | Example                                |
| ----------------------------- | -------------------------------------------------- | -------------------------------------- |
| Integration (cross-service)   | `events.integration.{service}.{aggregate}.{event}` | `events.integration.note.note.created` |
| Internal (inside one service) | `events.internal.{service}.*`                      | `events.internal.note.folder.renamed`  |

Every topic published or consumed by a Go service is a constant in that service's
`component/eventtopic.go`; no inline topic strings.

## Payloads

Cross-service payload types are defined once in `api/share/` (generated for Go and for
`@notopia-uit/api-share-gen`), never retyped per service. They cover document commits, note
created/updated/deleted, workspace deleted, workspace member added/removed, user workspace role
updates, user deleted and subscription changes.

See each service's page for its own topics, for example [Note events](../services/note/events).
