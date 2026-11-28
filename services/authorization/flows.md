---
url: /notopia/services/authorization/flows.md
---

# Authorization Flows

These flows describe the internal gRPC service. See the [overview](./index) for the
complete use-case and role-permission summaries.

## Check a permission

```mermaid
sequenceDiagram
    participant Caller as Calling service
    participant Auth as Authorization gRPC
    participant Casbin
    Caller->>Auth: HasWorkspacePermission or HasWorkspaceItemPermission
    Auth->>Casbin: Enforce(user, workspace, object class, action)
    Note over Casbin: Match workspace-scoped role and role policy
    Casbin-->>Auth: allow or deny
    Auth-->>Caller: has_permission
```

An enforcement error is returned as an error rather than a denial. For the combined
`GetUserWorkspaceItemPermissions` query, the service first requires workspace read,
then batch-checks item read, write and delete. Unlike the single permission checks,
the combined query returns an error when workspace read is denied.

## Replace workspace members

```mermaid
sequenceDiagram
    participant Caller as Calling service
    participant Auth as Authorization gRPC
    participant Casbin
    participant Kafka as Kafka publisher
    Caller->>Auth: UpdateWorkspaceMembers(user, workspace, complete list)
    Auth->>Casbin: Enforce workspace edit
    alt denied
        Casbin-->>Auth: deny
        Auth-->>Caller: permission error
    else allowed
        Casbin-->>Auth: allow
        Auth->>Casbin: Begin transaction and read current memberships
        Auth->>Casbin: Replace grouping policies
        Note over Auth: Compare old and new members
        opt membership changes
            Auth->>Kafka: Publish added, role-updated and removed events
        end
        Auth->>Casbin: Complete transaction
        Auth-->>Caller: success
    end
```

Publishing happens inside the Casbin transaction callback, through a direct Kafka
publisher, not a transactional outbox. Events are published individually, so partial
publication is possible and delivery is not atomic with the membership transaction.
On a publishing error the callback returns an error, but the update gRPC handler logs
that specific error and returns success. Do not interpret that response as a guarantee
that either all events were delivered or the membership transaction committed.

## Leave a workspace

```mermaid
flowchart TD
    A[LeaveWorkspace] --> B[Read workspace memberships in transaction]
    B --> C{User is a member?}
    C -- No --> D[Return permission error]
    C -- Yes --> E{User is the only owner?}
    E -- Yes --> F[Return only-owner error]
    E -- No --> G[Remove user grouping policy]
    G --> H[Publish member-removed event]
    H --> I{Publish succeeded?}
    I -- Yes --> J[Complete transaction and return success]
    I -- No --> K[Return error from transaction]
```

Unlike member replacement, leave propagates a publication error to the caller. The
direct publication is still not atomic with the Casbin transaction.

## Delete workspace authorization

`DeleteWorkspace` checks workspace delete permission before removing all grouping
policies for the workspace. It does not publish member-removal events in this handler;
workspace deletion and any related cleanup belong to the calling workflow.
