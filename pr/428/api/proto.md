---
url: /notopia/pr/428/api/proto.md
---
# Protocol Documentation

## Table of Contents

* [authorization.proto](#authorization-proto)
  * [CreateWorkspaceWithOwnerRequest](#authorization-CreateWorkspaceWithOwnerRequest)

  * [CreateWorkspaceWithOwnerResponse](#authorization-CreateWorkspaceWithOwnerResponse)

  * [DeleteWorkspaceRequest](#authorization-DeleteWorkspaceRequest)

  * [DeleteWorkspaceResponse](#authorization-DeleteWorkspaceResponse)

  * [GetUserWorkspaceItemPermissionsRequest](#authorization-GetUserWorkspaceItemPermissionsRequest)

  * [GetUserWorkspaceItemPermissionsResponse](#authorization-GetUserWorkspaceItemPermissionsResponse)

  * [GetUserWorkspacesRequest](#authorization-GetUserWorkspacesRequest)

  * [GetUserWorkspacesResponse](#authorization-GetUserWorkspacesResponse)

  * [GetWorkspaceMembersRequest](#authorization-GetWorkspaceMembersRequest)

  * [GetWorkspaceMembersResponse](#authorization-GetWorkspaceMembersResponse)

  * [HasWorkspaceItemPermissionRequest](#authorization-HasWorkspaceItemPermissionRequest)

  * [HasWorkspaceItemPermissionResponse](#authorization-HasWorkspaceItemPermissionResponse)

  * [HasWorkspacePermissionRequest](#authorization-HasWorkspacePermissionRequest)

  * [HasWorkspacePermissionResponse](#authorization-HasWorkspacePermissionResponse)

  * [LeaveWorkspaceRequest](#authorization-LeaveWorkspaceRequest)

  * [LeaveWorkspaceResponse](#authorization-LeaveWorkspaceResponse)

  * [UpdateWorkspaceMembersRequest](#authorization-UpdateWorkspaceMembersRequest)

  * [UpdateWorkspaceMembersResponse](#authorization-UpdateWorkspaceMembersResponse)

  * [UserWorkspace](#authorization-UserWorkspace)

  * [WorkspaceMember](#authorization-WorkspaceMember)

  * [WorkspaceItemPermission](#authorization-WorkspaceItemPermission)

  * [WorkspacePermission](#authorization-WorkspacePermission)

  * [WorkspaceRole](#authorization-WorkspaceRole)

  * [AuthorizationService](#authorization-AuthorizationService)

* [billing.proto](#billing-proto)
  * [BatchGetWorkspaceEntitlementsRequest](#billing-BatchGetWorkspaceEntitlementsRequest)

  * [BatchGetWorkspaceEntitlementsResponse](#billing-BatchGetWorkspaceEntitlementsResponse)

  * [BatchGetWorkspaceEntitlementsResponse.EntitlementsEntry](#billing-BatchGetWorkspaceEntitlementsResponse-EntitlementsEntry)

  * [BillingPlan](#billing-BillingPlan)

  * [Entitlements](#billing-Entitlements)

  * [Features](#billing-Features)

  * [GetWorkspaceEntitlementsRequest](#billing-GetWorkspaceEntitlementsRequest)

  * [GetWorkspaceEntitlementsResponse](#billing-GetWorkspaceEntitlementsResponse)

  * [GetWorkspacePayerRequest](#billing-GetWorkspacePayerRequest)

  * [GetWorkspacePayerResponse](#billing-GetWorkspacePayerResponse)

  * [Limits](#billing-Limits)

  * [ListPlansRequest](#billing-ListPlansRequest)

  * [ListPlansResponse](#billing-ListPlansResponse)

  * [PlanTier](#billing-PlanTier)

  * [SubscriptionStatus](#billing-SubscriptionStatus)

  * [BillingService](#billing-BillingService)

* [note.proto](#note-proto)
  * [BatchGetNoteWorkspacesRequest](#note-BatchGetNoteWorkspacesRequest)

  * [BatchGetNoteWorkspacesResponse](#note-BatchGetNoteWorkspacesResponse)

  * [BatchGetNoteWorkspacesResponse.WorkspaceIdsEntry](#note-BatchGetNoteWorkspacesResponse-WorkspaceIdsEntry)

  * [GetNoteNameRequest](#note-GetNoteNameRequest)

  * [GetNoteNameResponse](#note-GetNoteNameResponse)

  * [GetNoteRequest](#note-GetNoteRequest)

  * [GetNoteResponse](#note-GetNoteResponse)

  * [GetWorkspaceByNoteRequest](#note-GetWorkspaceByNoteRequest)

  * [GetWorkspaceByNoteResponse](#note-GetWorkspaceByNoteResponse)

  * [Note](#note-Note)

  * [Trashed](#note-Trashed)

  * [Workspace](#note-Workspace)

  * [TrashedBy](#note-TrashedBy)

  * [NoteService](#note-NoteService)

* [Scalar Value Types](#scalar-value-types)

## authorization.proto

### CreateWorkspaceWithOwnerRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| owner_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |

### CreateWorkspaceWithOwnerResponse

### DeleteWorkspaceRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |

### DeleteWorkspaceResponse

### GetUserWorkspaceItemPermissionsRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |

### GetUserWorkspaceItemPermissionsResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| can_read | [bool](#bool) |  |  |
| can_write | [bool](#bool) |  |  |
| can_delete | [bool](#bool) |  |  |

### GetUserWorkspacesRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |

### GetUserWorkspacesResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspaces | [UserWorkspace](#authorization-UserWorkspace) | repeated |  |

### GetWorkspaceMembersRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |

### GetWorkspaceMembersResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| members | [WorkspaceMember](#authorization-WorkspaceMember) | repeated |  |

### HasWorkspaceItemPermissionRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |
| permission | [WorkspaceItemPermission](#authorization-WorkspaceItemPermission) |  |  |

### HasWorkspaceItemPermissionResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| has_permission | [bool](#bool) |  |  |

### HasWorkspacePermissionRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |
| permission | [WorkspacePermission](#authorization-WorkspacePermission) |  |  |

### HasWorkspacePermissionResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| has_permission | [bool](#bool) |  |  |

### LeaveWorkspaceRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |

### LeaveWorkspaceResponse

### UpdateWorkspaceMembersRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |
| members | [WorkspaceMember](#authorization-WorkspaceMember) | repeated |  |

### UpdateWorkspaceMembersResponse

### UserWorkspace

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| role | [WorkspaceRole](#authorization-WorkspaceRole) |  |  |

### WorkspaceMember

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| role | [WorkspaceRole](#authorization-WorkspaceRole) |  |  |

### WorkspaceItemPermission

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKSPACE_ITEM_PERMISSION_UNSPECIFIED | 0 |  |
| WORKSPACE_ITEM_PERMISSION_READ | 1 |  |
| WORKSPACE_ITEM_PERMISSION_WRITE | 2 |  |
| WORKSPACE_ITEM_PERMISSION_DELETE | 3 |  |

### WorkspacePermission

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKSPACE_PERMISSION_UNSPECIFIED | 0 |  |
| WORKSPACE_PERMISSION_READ | 1 |  |
| WORKSPACE_PERMISSION_EDIT | 2 |  |
| WORKSPACE_PERMISSION_DELETE | 3 |  |
| WORKSPACE_PERMISSION_BILLING_READ | 4 |  |
| WORKSPACE_PERMISSION_BILLING_MANAGE | 5 |  |

### WorkspaceRole

| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKSPACE_ROLE_UNSPECIFIED | 0 |  |
| WORKSPACE_ROLE_OWNER | 1 |  |
| WORKSPACE_ROLE_EDITOR | 2 |  |
| WORKSPACE_ROLE_VIEWER | 3 |  |

### AuthorizationService

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetUserWorkspaces | [GetUserWorkspacesRequest](#authorization-GetUserWorkspacesRequest) | [GetUserWorkspacesResponse](#authorization-GetUserWorkspacesResponse) |  |
| CreateWorkspaceWithOwner | [CreateWorkspaceWithOwnerRequest](#authorization-CreateWorkspaceWithOwnerRequest) | [CreateWorkspaceWithOwnerResponse](#authorization-CreateWorkspaceWithOwnerResponse) |  |
| UpdateWorkspaceMembers | [UpdateWorkspaceMembersRequest](#authorization-UpdateWorkspaceMembersRequest) | [UpdateWorkspaceMembersResponse](#authorization-UpdateWorkspaceMembersResponse) |  |
| GetWorkspaceMembers | [GetWorkspaceMembersRequest](#authorization-GetWorkspaceMembersRequest) | [GetWorkspaceMembersResponse](#authorization-GetWorkspaceMembersResponse) |  |
| LeaveWorkspace | [LeaveWorkspaceRequest](#authorization-LeaveWorkspaceRequest) | [LeaveWorkspaceResponse](#authorization-LeaveWorkspaceResponse) |  |
| HasWorkspacePermission | [HasWorkspacePermissionRequest](#authorization-HasWorkspacePermissionRequest) | [HasWorkspacePermissionResponse](#authorization-HasWorkspacePermissionResponse) |  |
| HasWorkspaceItemPermission | [HasWorkspaceItemPermissionRequest](#authorization-HasWorkspaceItemPermissionRequest) | [HasWorkspaceItemPermissionResponse](#authorization-HasWorkspaceItemPermissionResponse) |  |
| GetUserWorkspaceItemPermissions | [GetUserWorkspaceItemPermissionsRequest](#authorization-GetUserWorkspaceItemPermissionsRequest) | [GetUserWorkspaceItemPermissionsResponse](#authorization-GetUserWorkspaceItemPermissionsResponse) |  |
| DeleteWorkspace | [DeleteWorkspaceRequest](#authorization-DeleteWorkspaceRequest) | [DeleteWorkspaceResponse](#authorization-DeleteWorkspaceResponse) |  |

## billing.proto

### BatchGetWorkspaceEntitlementsRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_ids | [string](#string) | repeated |  |

### BatchGetWorkspaceEntitlementsResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entitlements | [BatchGetWorkspaceEntitlementsResponse.EntitlementsEntry](#billing-BatchGetWorkspaceEntitlementsResponse-EntitlementsEntry) | repeated |  |

### BatchGetWorkspaceEntitlementsResponse.EntitlementsEntry

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Entitlements](#billing-Entitlements) |  |  |

### BillingPlan

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tier | [PlanTier](#billing-PlanTier) |  |  |
| features | [Features](#billing-Features) |  |  |
| limits | [Limits](#billing-Limits) |  |  |

### Entitlements

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_id | [string](#string) |  |  |
| plan | [PlanTier](#billing-PlanTier) |  |  |
| status | [SubscriptionStatus](#billing-SubscriptionStatus) |  |  |
| access_until | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional |  |
| features | [Features](#billing-Features) |  |  |
| limits | [Limits](#billing-Limits) |  |  |

### Features

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auto_commit | [bool](#bool) | optional |  |
| workspace_graph | [bool](#bool) | optional |  |

### GetWorkspaceEntitlementsRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_id | [string](#string) |  |  |

### GetWorkspaceEntitlementsResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entitlements | [Entitlements](#billing-Entitlements) |  |  |

### GetWorkspacePayerRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_id | [string](#string) |  |  |

### GetWorkspacePayerResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payer_id | [string](#string) | optional |  |
| requires_payer | [bool](#bool) |  |  |

### Limits

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| members | [int64](#int64) | optional |  |
| published_notes | [int64](#int64) | optional |  |
| note_graph_max_depth | [int64](#int64) | optional |  |
| auto_commit_min_interval_seconds | [int64](#int64) | optional |  |
| revision_history_days | [int64](#int64) | optional |  |
| trash_retention_days | [int64](#int64) | optional |  |
| ai_credits_monthly | [int64](#int64) | optional |  |

### ListPlansRequest

### ListPlansResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| plans | [BillingPlan](#billing-BillingPlan) | repeated |  |

### PlanTier

| Name | Number | Description |
| ---- | ------ | ----------- |
| PLAN_TIER_UNSPECIFIED | 0 |  |
| PLAN_TIER_FREE | 1 |  |
| PLAN_TIER_PLUS | 2 |  |
| PLAN_TIER_PRO | 3 |  |

### SubscriptionStatus

| Name | Number | Description |
| ---- | ------ | ----------- |
| SUBSCRIPTION_STATUS_UNSPECIFIED | 0 |  |
| SUBSCRIPTION_STATUS_NONE | 1 |  |
| SUBSCRIPTION_STATUS_INCOMPLETE | 2 |  |
| SUBSCRIPTION_STATUS_ACTIVE | 3 |  |
| SUBSCRIPTION_STATUS_PAST_DUE | 4 |  |
| SUBSCRIPTION_STATUS_UNPAID | 5 |  |
| SUBSCRIPTION_STATUS_CANCELED | 6 |  |
| SUBSCRIPTION_STATUS_INCOMPLETE_EXPIRED | 7 |  |

### BillingService

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetWorkspaceEntitlements | [GetWorkspaceEntitlementsRequest](#billing-GetWorkspaceEntitlementsRequest) | [GetWorkspaceEntitlementsResponse](#billing-GetWorkspaceEntitlementsResponse) |  |
| BatchGetWorkspaceEntitlements | [BatchGetWorkspaceEntitlementsRequest](#billing-BatchGetWorkspaceEntitlementsRequest) | [BatchGetWorkspaceEntitlementsResponse](#billing-BatchGetWorkspaceEntitlementsResponse) |  |
| GetWorkspacePayer | [GetWorkspacePayerRequest](#billing-GetWorkspacePayerRequest) | [GetWorkspacePayerResponse](#billing-GetWorkspacePayerResponse) |  |
| ListPlans | [ListPlansRequest](#billing-ListPlansRequest) | [ListPlansResponse](#billing-ListPlansResponse) |  |

## note.proto

### BatchGetNoteWorkspacesRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| note_ids | [string](#string) | repeated |  |

### BatchGetNoteWorkspacesResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_ids | [BatchGetNoteWorkspacesResponse.WorkspaceIdsEntry](#note-BatchGetNoteWorkspacesResponse-WorkspaceIdsEntry) | repeated |  |

### BatchGetNoteWorkspacesResponse.WorkspaceIdsEntry

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |

### GetNoteNameRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |

### GetNoteNameResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |

### GetNoteRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| exclude_trashed | [bool](#bool) |  |  |
| user_id | [string](#string) |  |  |

### GetNoteResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| note | [Note](#note-Note) |  |  |

### GetWorkspaceByNoteRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| note_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |

### GetWorkspaceByNoteResponse

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace | [Workspace](#note-Workspace) |  |  |

### Note

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| icon | [string](#string) | optional |  |
| folder_id | [string](#string) |  |  |
| tags | [string](#string) | repeated |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| trashed | [Trashed](#note-Trashed) | optional |  |

### Trashed

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| by | [TrashedBy](#note-TrashedBy) |  |  |
| at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |

### Workspace

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| name | [string](#string) |  |  |

### TrashedBy

| Name | Number | Description |
| ---- | ------ | ----------- |
| TRASHED_BY_UNSPECIFIED | 0 |  |
| TRASHED_BY_PURPOSE | 1 |  |
| TRASHED_BY_PARENT | 2 |  |

### NoteService

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetNoteName | [GetNoteNameRequest](#note-GetNoteNameRequest) | [GetNoteNameResponse](#note-GetNoteNameResponse) | NOTE: maybe this will be replaced with batch get |
| GetNote | [GetNoteRequest](#note-GetNoteRequest) | [GetNoteResponse](#note-GetNoteResponse) |  |
| GetWorkspaceByNote | [GetWorkspaceByNoteRequest](#note-GetWorkspaceByNoteRequest) | [GetWorkspaceByNoteResponse](#note-GetWorkspaceByNoteResponse) | NOTE: what, where is this used? |
| BatchGetNoteWorkspaces | [BatchGetNoteWorkspacesRequest](#note-BatchGetNoteWorkspacesRequest) | [BatchGetNoteWorkspacesResponse](#note-BatchGetNoteWorkspacesResponse) |  |

## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
|  double |  | double | double | float | float64 | double | float | Float |
|  float |  | float | float | float | float32 | float | float | Float |
|  int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
|  int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
|  uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
|  uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
|  sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
|  sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
|  fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
|  fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
|  sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
|  sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
|  bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
|  string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
|  bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | \[]byte | ByteString | string | String (ASCII-8BIT) |
