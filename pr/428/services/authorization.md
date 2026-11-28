---
url: /notopia/pr/428/services/authorization.md
---

# Authorization Service

Go service that answers permission checks (Casbin policies, workspace memberships and roles)
over gRPC for the other services.

* [Domain](./domain): class diagrams
* [Data](./data): database schema
* [Flows](./flows): permission checks and membership changes

## Use cases

Authorization exposes an internal gRPC API; callers supply the user and workspace IDs.
It stores workspace-scoped role assignments, not workspace metadata or individual note
ACLs. The calling service is responsible for identifying the user and enforcing the
result of a check.

| Use case                  | RPC                                                    | Behavior                                                                                                                                  |
| ------------------------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Create initial membership | `CreateWorkspaceWithOwner`                             | Adds the supplied owner-to-workspace role assignment; a duplicate assignment is an error.                                                 |
| List a user's workspaces  | `GetUserWorkspaces`                                    | Returns workspace IDs and roles from that user's grouping policies.                                                                       |
| List members              | `GetWorkspaceMembers`                                  | Requires workspace read permission and returns user IDs and roles.                                                                        |
| Replace members           | `UpdateWorkspaceMembers`                               | Requires workspace edit permission and replaces the complete membership list, generating events for additions, role changes and removals. |
| Leave a workspace         | `LeaveWorkspace`                                       | Removes the caller's membership, except when the caller is the only owner.                                                                |
| Delete authorization data | `DeleteWorkspace`                                      | Requires workspace delete permission and removes all membership policies for that workspace, not the workspace itself.                    |
| Check access              | `HasWorkspacePermission`, `HasWorkspaceItemPermission` | Return a boolean for a workspace action or a workspace-wide item action.                                                                  |
| Get item capabilities     | `GetUserWorkspaceItemPermissions`                      | Requires workspace read permission, then returns item read, write and delete decisions.                                                   |

The replacement request must contain at least one member. Unlike `LeaveWorkspace`, its
handler does not check that an owner remains. These operations do not create or delete
the corresponding workspace in another service.

## Roles and permissions

The seeded policy grants the following actions:

| Role   | Workspace                                        | Workspace items     |
| ------ | ------------------------------------------------ | ------------------- |
| Owner  | read, edit, delete, billing read, billing manage | read, write, delete |
| Editor | read, billing read                               | read, write, delete |
| Viewer | read, billing read                               | read                |

Casbin's `g` policies associate a user with a role in a workspace. Static `p` policies
grant actions to roles; `g2` maps note and folder objects to the `workspace_item` class.
The item-check RPCs check the class for the whole workspace, not a particular item ID.

The static policy is seeded separately with `just seed-authorization`; development
memberships are loaded separately with `just fixtures-authorization`. At startup,
policy drift is logged and fails the policy health check; the running enforcer does not
watch the database, so restart it after correcting the seed.

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
  grpc: gRPC
}

application: Application

infrastructure: Infrastructure {
  casbin_enforcer: Casbin Enforcer
  persistence: Persistence/GORM
}

controllers -> application
application -> infrastructure
```
