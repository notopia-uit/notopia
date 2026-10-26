# The Event Processor That Would Have Eaten Every Message

_Why the billing service's Kafka consumers don't use Watermill's CQRS `EventProcessor`, even though the note service does, and the one-line metadata convention that makes the difference._

Written with AI while building `internal/billing/controller/event` (billing-03, Task 10).

---

## The Setup

Billing has to react to three things that happen in other services:

| Topic                                                          | Publisher          | Billing reaction                                                      |
| -------------------------------------------------------------- | ------------------ | --------------------------------------------------------------------- |
| `events.integration.note.workspace.deleted`                    | note (Go)          | Cancel the subscription now, refund the unused part                   |
| `events.integration.authorization.user_workspace_role_updated` | authorization (Go) | If the payer lost the owner role: payer missing, cancel at period end |
| `events.integration.authorization.workspace_member_removed`    | authorization (Go) | Same as above                                                         |

The payload types live once in `api/share/` and are generated into `pkg/api/share`. Our services are on Watermill, and the note service already consumes Kafka. Copying its pattern was the obvious move.

Note's event controller uses Watermill's [CQRS component](https://watermill.io/docs/cqrs/), which looks great:

```go
eventProcessor.AddHandler(cqrs.NewEventHandler(
    "NotifyWorkspaceRenamedHandler",
    e.app.Events.NotifyWorkspaceRenamed.Handle, // func(ctx, *WorkspaceRenamed) error
))
```

You get typed handlers, no `json.Unmarshal` boilerplate, and a topic derived from the Go type. The obvious question in review: _why doesn't billing do this too?_

---

## How the EventProcessor Actually Routes

The EventProcessor doesn't just decode the payload into the handler's type. Before it calls your handler, it checks the message's identity, in `components/cqrs/event_processor.go` (v1.5.2):

```go
expectedEventName := p.config.Marshaler.Name(initEvent)

return func(msg *message.Message) error {
    messageEventName := p.config.Marshaler.NameFromMessage(msg)

    if messageEventName != expectedEventName {
        if !p.config.AckOnUnknownEvent {
            return fmt.Errorf("received unexpected event type %s, expected %s", ...)
        } else {
            // log "Received different event type than expected" and ack
            return nil
        }
    }
    // ... unmarshal + call handler
}
```

And `JSONMarshaler.NameFromMessage` is:

```go
func (m JSONMarshaler) NameFromMessage(msg *message.Message) string {
    return msg.Metadata.Get("name")
}
```

So the processor routes by a **`name` metadata header**. The topic alone isn't enough: the message must also say what it is. `JSONMarshaler.Marshal` sets that header, which means it only works when the **publisher also uses the CQRS marshaler** (an `EventBus`, or an outbox that marshals with it).

---

## Who Sets `name`? Nobody We Consume From

| Publisher                                     | How it publishes                      | Sets `name`? |
| --------------------------------------------- | ------------------------------------- | ------------ |
| note domain events (its own outbox)           | `cqrs.JSONMarshaler`                  | ✅           |
| note integration events (`workspace.deleted`) | `json.Marshal` + `message.NewMessage` | ❌           |
| authorization `IntegrationPublisher`          | `json.Marshal` + `message.NewMessage` | ❌           |
| document (NestJS)                             | `@nestjs/microservices` Kafka client  | ❌           |

With the default config (`AckOnUnknownEvent: true`, kept for backwards compatibility), every billing handler would have **acked and dropped every message** with one log line at info level. Workspace deletions would never cancel the subscription and payer removals would never be noticed. Nothing crashes and no test fails. You would find out when a customer gets charged for a deleted workspace.

With `AckOnUnknownEvent: false` it's loud instead: an error, a nack, and the same message redelivered forever, blocking its partition.

---

## Note Already Knew This

Read note's `controller/event/event.go` closely and the split is already there:

- **Its own domain events** (outbox, CQRS marshaler, `name` set) → `eventProcessor.AddHandler(...)`
- **`document.committed` from NestJS** (no `name`) → plain `router.AddConsumerHandler(...)` with a hand-written `json.Unmarshal`

That wasn't an accident: the processor simply can't see foreign messages. Billing has no domain-event outbox at all (its only outbound notification is best-effort), so **every** billing consumer is in the foreign-message case.

---

## The Second Gap: One Consumer Group, Many Topics

Note's controller carries this comment:

```go
// Be careful if we use single subcriber, they will steal each other message
```

One Watermill Kafka subscriber means one `ConsumerGroup`. If three handlers each call `Subscribe` on different topics under the same group, Sarama runs three group members that each subscribe to a different topic. Every member joining or leaving triggers a rebalance of the whole group, which pauses consumption on _all_ three topics, and a slow handler on one topic can stall the group's session for the others. Note hit this as "stealing" in practice. Billing therefore creates **one subscriber per handler**, with group `billing-service.<handler>`:

```go
for _, h := range handlers {
    subscriber, err := kafka.NewSubscriber(kafka.SubscriberConfig{
        ConsumerGroup: kafkaCfg.ConsumerGroup + "." + h.name,
        // ...
    }, logger)
    router.AddConsumerHandler(h.name, h.topic, subscriber, h.handler)
}
```

---

## Getting the Typed Handlers Anyway

What we actually wanted from the processor was the typing, not the routing. A small generic helper gives us that without the name check:

```go
// decode turns a typed handler into a router handler. A payload that does not decode is
// logged and acked: retrying cannot fix it, and a nack would block the partition.
func decode[T any](fn func(ctx context.Context, ev *T) error) message.NoPublishHandlerFunc {
    return func(msg *message.Message) error {
        var ev T
        if err := json.Unmarshal(msg.Payload, &ev); err != nil {
            slog.ErrorContext(msg.Context(), "dropping undecodable event", ...)
            return nil
        }
        return fn(msg.Context(), &ev)
    }
}
```

Handlers now read like the CQRS ones:

```go
func (e *Event) roleUpdated(ctx context.Context, ev *share.UserWorkspaceRoleUpdatedEvent) error {
    if ev.Role == share.Owner || ev.WorkspaceId == nil {
        return nil
    }
    return e.app.Cmds.HandlePayerRemoved.Handle(ctx, &app.HandlePayerRemoved{...})
}

router.AddConsumerHandler("role-updated", component.TopicRoleUpdated, sub, decode(e.roleUpdated))
```

There's one deliberate difference from note: a **malformed payload is logged and acked**, not returned as an error. A JSON error is permanent, and Watermill's Kafka subscriber redelivers a nacked message until it succeeds. Since billing has no poison queue, returning the error would freeze the partition behind one bad message. Errors from the command itself (Stripe down, database down) are still returned, so the retry middleware and redelivery handle them.

---

## If We Ever Want the EventProcessor for Integration Events

It's possible, but it's a cross-service contract change, not a billing change:

1. Every publisher sets `name` metadata: both Go integration publishers, plus the NestJS producers (as a Kafka header, which Watermill maps to metadata).
2. The value must equal what the consumer's marshaler generates. `JSONMarshaler` defaults to the Go struct name (`cqrs.FullyQualifiedStructName` or `StructName`), so a TypeScript producer would have to emit `share.UserWorkspaceRoleUpdatedEvent`. That leaks Go naming into the contract, unless we set `GenerateName` on both sides to the topic name.
3. Document it in `api/share/` next to the payload schemas, because a publisher that forgets it fails silently (see above).

Until a consumer needs to fan out on event type inside one topic, the topic already is the type, and the plain router plus `decode[T]` is the simpler contract.

---

## Takeaways

- Watermill's CQRS `EventProcessor` routes on the **`name` metadata**, not just the topic. Messages from publishers that don't use the CQRS marshaler are acked and dropped by default.
- Use the processor for events your service publishes itself through the CQRS marshaler. Use `router.AddConsumerHandler` for anything that crosses a service boundary.
- One consumer group per topic handler.
- Decode errors are permanent: log and ack. Handler errors are transient: return them.
