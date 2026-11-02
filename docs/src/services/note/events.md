---
title: Events
order: 5
---

# Note Events

Topic constants live in `internal/note/component/eventtopic.go`. See
[Messaging](../../architecture/messaging) for the outbox mechanics.

## Published: integration events

Prefix `events.integration.note.`, written through the transactional outbox.

| Topic               | Notes                                                                                                                           |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `note.created`      |                                                                                                                                 |
| `note.updated`      |                                                                                                                                 |
| `note.deleted`      | Every permanent delete (folder, note, trash) deletes the affected notes through the repo first, so each one emits its own event |
| `workspace.deleted` |                                                                                                                                 |

## Internal: domain events

Prefix `events.internal.note.`, drained from aggregates on `Save`.

| Aggregate | Topics                                                                                                                                                           |
| --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| folder    | `created`, `deleted`, `renamed`, `icon_changed`, `moved`, `trashed`, `restored`, `permanently_deleted`                                                           |
| note      | `created`, `deleted`, `renamed`, `icon_changed`, `tags_changed`, `size_changed`, `moved`, `outgoing_links_changed`, `trashed`, `restored`, `permanently_deleted` |
| workspace | `renamed`, `slug_changed`, `deleted`                                                                                                                             |

A `poison` topic under the same prefix receives messages that failed handling.

## Consumed

| Topic                                             | Handler                                                                                   |
| ------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `events.integration.document.document.committed`  | Updates note size, tags and outgoing links                                                |
| `events.integration.billing.subscription.changed` | Publishes `WorkspaceBillingChangedEvent {workspaceId, plan}` to the workspace SSE channel |
| `events.internal.note.folder.*` and `note.*`      | Notify workspace item updates over SSE                                                    |
