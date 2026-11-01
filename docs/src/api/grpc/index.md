# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [authorization.proto](#authorization-proto)
    - [CreateWorkspaceWithOwnerRequest](#authorization-CreateWorkspaceWithOwnerRequest)
    - [CreateWorkspaceWithOwnerResponse](#authorization-CreateWorkspaceWithOwnerResponse)
    - [DeleteWorkspaceRequest](#authorization-DeleteWorkspaceRequest)
    - [DeleteWorkspaceResponse](#authorization-DeleteWorkspaceResponse)
    - [GetUserWorkspaceItemPermissionsRequest](#authorization-GetUserWorkspaceItemPermissionsRequest)
    - [GetUserWorkspaceItemPermissionsResponse](#authorization-GetUserWorkspaceItemPermissionsResponse)
    - [GetUserWorkspacesRequest](#authorization-GetUserWorkspacesRequest)
    - [GetUserWorkspacesResponse](#authorization-GetUserWorkspacesResponse)
    - [GetWorkspaceMembersRequest](#authorization-GetWorkspaceMembersRequest)
    - [GetWorkspaceMembersResponse](#authorization-GetWorkspaceMembersResponse)
    - [HasWorkspaceItemPermissionRequest](#authorization-HasWorkspaceItemPermissionRequest)
    - [HasWorkspaceItemPermissionResponse](#authorization-HasWorkspaceItemPermissionResponse)
    - [HasWorkspacePermissionRequest](#authorization-HasWorkspacePermissionRequest)
    - [HasWorkspacePermissionResponse](#authorization-HasWorkspacePermissionResponse)
    - [LeaveWorkspaceRequest](#authorization-LeaveWorkspaceRequest)
    - [LeaveWorkspaceResponse](#authorization-LeaveWorkspaceResponse)
    - [UpdateWorkspaceMembersRequest](#authorization-UpdateWorkspaceMembersRequest)
    - [UpdateWorkspaceMembersResponse](#authorization-UpdateWorkspaceMembersResponse)
    - [UserWorkspace](#authorization-UserWorkspace)
    - [WorkspaceMember](#authorization-WorkspaceMember)
  
    - [WorkspaceItemPermission](#authorization-WorkspaceItemPermission)
    - [WorkspacePermission](#authorization-WorkspacePermission)
    - [WorkspaceRole](#authorization-WorkspaceRole)
  
    - [AuthorizationService](#authorization-AuthorizationService)
  
- [billing.proto](#billing-proto)
    - [BatchGetWorkspaceEntitlementsRequest](#billing-BatchGetWorkspaceEntitlementsRequest)
    - [BatchGetWorkspaceEntitlementsResponse](#billing-BatchGetWorkspaceEntitlementsResponse)
    - [BatchGetWorkspaceEntitlementsResponse.EntitlementsEntry](#billing-BatchGetWorkspaceEntitlementsResponse-EntitlementsEntry)
    - [BillingPlan](#billing-BillingPlan)
    - [BillingPlan.FeaturesEntry](#billing-BillingPlan-FeaturesEntry)
    - [BillingPlan.LimitsEntry](#billing-BillingPlan-LimitsEntry)
    - [Entitlements](#billing-Entitlements)
    - [Entitlements.FeaturesEntry](#billing-Entitlements-FeaturesEntry)
    - [Entitlements.LimitsEntry](#billing-Entitlements-LimitsEntry)
    - [GetWorkspaceEntitlementsRequest](#billing-GetWorkspaceEntitlementsRequest)
    - [GetWorkspaceEntitlementsResponse](#billing-GetWorkspaceEntitlementsResponse)
    - [GetWorkspacePayerRequest](#billing-GetWorkspacePayerRequest)
    - [GetWorkspacePayerResponse](#billing-GetWorkspacePayerResponse)
    - [ListPlansRequest](#billing-ListPlansRequest)
    - [ListPlansResponse](#billing-ListPlansResponse)
  
    - [PlanTier](#billing-PlanTier)
    - [SubscriptionStatus](#billing-SubscriptionStatus)
  
    - [BillingService](#billing-BillingService)
  
- [note.proto](#note-proto)
    - [BatchGetNoteWorkspacesRequest](#note-BatchGetNoteWorkspacesRequest)
    - [BatchGetNoteWorkspacesResponse](#note-BatchGetNoteWorkspacesResponse)
    - [BatchGetNoteWorkspacesResponse.WorkspaceIdsEntry](#note-BatchGetNoteWorkspacesResponse-WorkspaceIdsEntry)
    - [GetNoteNameRequest](#note-GetNoteNameRequest)
    - [GetNoteNameResponse](#note-GetNoteNameResponse)
    - [GetNoteRequest](#note-GetNoteRequest)
    - [GetNoteResponse](#note-GetNoteResponse)
    - [GetWorkspaceByNoteRequest](#note-GetWorkspaceByNoteRequest)
    - [GetWorkspaceByNoteResponse](#note-GetWorkspaceByNoteResponse)
    - [Note](#note-Note)
    - [Trashed](#note-Trashed)
    - [Workspace](#note-Workspace)
  
    - [TrashedBy](#note-TrashedBy)
  
    - [NoteService](#note-NoteService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="authorization-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## authorization.proto



<a name="authorization-CreateWorkspaceWithOwnerRequest"></a>

### CreateWorkspaceWithOwnerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| owner_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |






<a name="authorization-CreateWorkspaceWithOwnerResponse"></a>

### CreateWorkspaceWithOwnerResponse







<a name="authorization-DeleteWorkspaceRequest"></a>

### DeleteWorkspaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |






<a name="authorization-DeleteWorkspaceResponse"></a>

### DeleteWorkspaceResponse







<a name="authorization-GetUserWorkspaceItemPermissionsRequest"></a>

### GetUserWorkspaceItemPermissionsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |






<a name="authorization-GetUserWorkspaceItemPermissionsResponse"></a>

### GetUserWorkspaceItemPermissionsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| can_read | [bool](#bool) |  |  |
| can_write | [bool](#bool) |  |  |
| can_delete | [bool](#bool) |  |  |






<a name="authorization-GetUserWorkspacesRequest"></a>

### GetUserWorkspacesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |






<a name="authorization-GetUserWorkspacesResponse"></a>

### GetUserWorkspacesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspaces | [UserWorkspace](#authorization-UserWorkspace) | repeated |  |






<a name="authorization-GetWorkspaceMembersRequest"></a>

### GetWorkspaceMembersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |






<a name="authorization-GetWorkspaceMembersResponse"></a>

### GetWorkspaceMembersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| members | [WorkspaceMember](#authorization-WorkspaceMember) | repeated |  |






<a name="authorization-HasWorkspaceItemPermissionRequest"></a>

### HasWorkspaceItemPermissionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |
| permission | [WorkspaceItemPermission](#authorization-WorkspaceItemPermission) |  |  |






<a name="authorization-HasWorkspaceItemPermissionResponse"></a>

### HasWorkspaceItemPermissionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| has_permission | [bool](#bool) |  |  |






<a name="authorization-HasWorkspacePermissionRequest"></a>

### HasWorkspacePermissionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| member_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |
| permission | [WorkspacePermission](#authorization-WorkspacePermission) |  |  |






<a name="authorization-HasWorkspacePermissionResponse"></a>

### HasWorkspacePermissionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| has_permission | [bool](#bool) |  |  |






<a name="authorization-LeaveWorkspaceRequest"></a>

### LeaveWorkspaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |






<a name="authorization-LeaveWorkspaceResponse"></a>

### LeaveWorkspaceResponse







<a name="authorization-UpdateWorkspaceMembersRequest"></a>

### UpdateWorkspaceMembersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| workspace_id | [string](#string) |  |  |
| members | [WorkspaceMember](#authorization-WorkspaceMember) | repeated |  |






<a name="authorization-UpdateWorkspaceMembersResponse"></a>

### UpdateWorkspaceMembersResponse







<a name="authorization-UserWorkspace"></a>

### UserWorkspace



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| role | [WorkspaceRole](#authorization-WorkspaceRole) |  |  |






<a name="authorization-WorkspaceMember"></a>

### WorkspaceMember



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| role | [WorkspaceRole](#authorization-WorkspaceRole) |  |  |





 


<a name="authorization-WorkspaceItemPermission"></a>

### WorkspaceItemPermission


| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKSPACE_ITEM_PERMISSION_UNSPECIFIED | 0 |  |
| WORKSPACE_ITEM_PERMISSION_READ | 1 |  |
| WORKSPACE_ITEM_PERMISSION_WRITE | 2 |  |
| WORKSPACE_ITEM_PERMISSION_DELETE | 3 |  |



<a name="authorization-WorkspacePermission"></a>

### WorkspacePermission


| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKSPACE_PERMISSION_UNSPECIFIED | 0 |  |
| WORKSPACE_PERMISSION_READ | 1 |  |
| WORKSPACE_PERMISSION_EDIT | 2 |  |
| WORKSPACE_PERMISSION_DELETE | 3 |  |
| WORKSPACE_PERMISSION_BILLING_READ | 4 |  |
| WORKSPACE_PERMISSION_BILLING_MANAGE | 5 |  |



<a name="authorization-WorkspaceRole"></a>

### WorkspaceRole


| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKSPACE_ROLE_UNSPECIFIED | 0 |  |
| WORKSPACE_ROLE_OWNER | 1 |  |
| WORKSPACE_ROLE_EDITOR | 2 |  |
| WORKSPACE_ROLE_VIEWER | 3 |  |


 

 


<a name="authorization-AuthorizationService"></a>

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

 



<a name="billing-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## billing.proto



<a name="billing-BatchGetWorkspaceEntitlementsRequest"></a>

### BatchGetWorkspaceEntitlementsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_ids | [string](#string) | repeated |  |






<a name="billing-BatchGetWorkspaceEntitlementsResponse"></a>

### BatchGetWorkspaceEntitlementsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entitlements | [BatchGetWorkspaceEntitlementsResponse.EntitlementsEntry](#billing-BatchGetWorkspaceEntitlementsResponse-EntitlementsEntry) | repeated |  |






<a name="billing-BatchGetWorkspaceEntitlementsResponse-EntitlementsEntry"></a>

### BatchGetWorkspaceEntitlementsResponse.EntitlementsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Entitlements](#billing-Entitlements) |  |  |






<a name="billing-BillingPlan"></a>

### BillingPlan



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tier | [PlanTier](#billing-PlanTier) |  |  |
| features | [BillingPlan.FeaturesEntry](#billing-BillingPlan-FeaturesEntry) | repeated |  |
| limits | [BillingPlan.LimitsEntry](#billing-BillingPlan-LimitsEntry) | repeated |  |






<a name="billing-BillingPlan-FeaturesEntry"></a>

### BillingPlan.FeaturesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="billing-BillingPlan-LimitsEntry"></a>

### BillingPlan.LimitsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="billing-Entitlements"></a>

### Entitlements
features/limits stay string-keyed (proto map keys cannot be enums): keys are FeatureKey/LimitKey names.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_id | [string](#string) |  |  |
| plan | [PlanTier](#billing-PlanTier) |  |  |
| status | [SubscriptionStatus](#billing-SubscriptionStatus) |  |  |
| access_until | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional |  |
| features | [Entitlements.FeaturesEntry](#billing-Entitlements-FeaturesEntry) | repeated |  |
| limits | [Entitlements.LimitsEntry](#billing-Entitlements-LimitsEntry) | repeated |  |






<a name="billing-Entitlements-FeaturesEntry"></a>

### Entitlements.FeaturesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="billing-Entitlements-LimitsEntry"></a>

### Entitlements.LimitsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="billing-GetWorkspaceEntitlementsRequest"></a>

### GetWorkspaceEntitlementsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_id | [string](#string) |  |  |






<a name="billing-GetWorkspaceEntitlementsResponse"></a>

### GetWorkspaceEntitlementsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entitlements | [Entitlements](#billing-Entitlements) |  |  |






<a name="billing-GetWorkspacePayerRequest"></a>

### GetWorkspacePayerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_id | [string](#string) |  |  |






<a name="billing-GetWorkspacePayerResponse"></a>

### GetWorkspacePayerResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| payer_id | [string](#string) | optional |  |
| requires_payer | [bool](#bool) |  |  |






<a name="billing-ListPlansRequest"></a>

### ListPlansRequest







<a name="billing-ListPlansResponse"></a>

### ListPlansResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| plans | [BillingPlan](#billing-BillingPlan) | repeated |  |





 


<a name="billing-PlanTier"></a>

### PlanTier


| Name | Number | Description |
| ---- | ------ | ----------- |
| PLAN_TIER_UNSPECIFIED | 0 |  |
| PLAN_TIER_FREE | 1 |  |
| PLAN_TIER_PLUS | 2 |  |
| PLAN_TIER_PRO | 3 |  |



<a name="billing-SubscriptionStatus"></a>

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


 

 


<a name="billing-BillingService"></a>

### BillingService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetWorkspaceEntitlements | [GetWorkspaceEntitlementsRequest](#billing-GetWorkspaceEntitlementsRequest) | [GetWorkspaceEntitlementsResponse](#billing-GetWorkspaceEntitlementsResponse) |  |
| BatchGetWorkspaceEntitlements | [BatchGetWorkspaceEntitlementsRequest](#billing-BatchGetWorkspaceEntitlementsRequest) | [BatchGetWorkspaceEntitlementsResponse](#billing-BatchGetWorkspaceEntitlementsResponse) |  |
| GetWorkspacePayer | [GetWorkspacePayerRequest](#billing-GetWorkspacePayerRequest) | [GetWorkspacePayerResponse](#billing-GetWorkspacePayerResponse) |  |
| ListPlans | [ListPlansRequest](#billing-ListPlansRequest) | [ListPlansResponse](#billing-ListPlansResponse) |  |

 



<a name="note-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## note.proto



<a name="note-BatchGetNoteWorkspacesRequest"></a>

### BatchGetNoteWorkspacesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| note_ids | [string](#string) | repeated |  |






<a name="note-BatchGetNoteWorkspacesResponse"></a>

### BatchGetNoteWorkspacesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace_ids | [BatchGetNoteWorkspacesResponse.WorkspaceIdsEntry](#note-BatchGetNoteWorkspacesResponse-WorkspaceIdsEntry) | repeated |  |






<a name="note-BatchGetNoteWorkspacesResponse-WorkspaceIdsEntry"></a>

### BatchGetNoteWorkspacesResponse.WorkspaceIdsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="note-GetNoteNameRequest"></a>

### GetNoteNameRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="note-GetNoteNameResponse"></a>

### GetNoteNameResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="note-GetNoteRequest"></a>

### GetNoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| exclude_trashed | [bool](#bool) |  |  |
| user_id | [string](#string) |  |  |






<a name="note-GetNoteResponse"></a>

### GetNoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| note | [Note](#note-Note) |  |  |






<a name="note-GetWorkspaceByNoteRequest"></a>

### GetWorkspaceByNoteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| note_id | [string](#string) |  |  |
| user_id | [string](#string) |  |  |






<a name="note-GetWorkspaceByNoteResponse"></a>

### GetWorkspaceByNoteResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace | [Workspace](#note-Workspace) |  |  |






<a name="note-Note"></a>

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






<a name="note-Trashed"></a>

### Trashed



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| by | [TrashedBy](#note-TrashedBy) |  |  |
| at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="note-Workspace"></a>

### Workspace



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| slug | [string](#string) |  |  |
| name | [string](#string) |  |  |





 


<a name="note-TrashedBy"></a>

### TrashedBy


| Name | Number | Description |
| ---- | ------ | ----------- |
| TRASHED_BY_UNSPECIFIED | 0 |  |
| TRASHED_BY_PURPOSE | 1 |  |
| TRASHED_BY_PARENT | 2 |  |


 

 


<a name="note-NoteService"></a>

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
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

