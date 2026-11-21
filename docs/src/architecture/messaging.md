---
title: Messaging
order: 2
---

# Messaging

Services communicate asynchronously through Redpanda (Kafka API).

## Transactional outbox

The Note and Billing services use a Watermill transactional outbox for aggregate events so
that a domain write and its outbox messages commit together. The flow in the Note service:

1. The repository `Save` drains domain events and publishes them with the Watermill
   `forwarder` publisher, which writes the message into an outbox table in the same Postgres
   transaction as the aggregate change.
2. The outbox table (default name `eventsToForward`, config `outbox_table_name`) follows the
   Watermill SQL queue schema (deleted on ack) and is created by a goose migration.
3. A Watermill SQL subscriber polls the table (every second) and the forwarder re-publishes each
   message to its real topic on Redpanda.

Authorization publishes membership events directly rather than through a transactional
outbox. NestJS services publish and consume directly with `@nestjs/microservices`; document
commit events are not atomic with the database transaction. See the
[authorization](/services/authorization/flows) and [document](/services/document/flows)
flows for delivery caveats.

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
