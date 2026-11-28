# SSE workspace updates

`GET /note/workspaces/{workspaceId}/events`

## Servers

- `http://api.notopia.localhost`

## Authorization

Use one of the following options:

- Option 1:
  - `note_Oauth2`: OAuth 2.0. Required scopes: `openid`. OAuth2 Security Scheme
    - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
    - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/
- Option 2:
  - `note_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). Required scopes: `openid`. OpenID Connect Security Scheme

## Parameters

### Path parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `workspaceId` | `string<uuid>` | Yes | Example: `00000000-0000-4000-8000-000000000110`. |

## Responses

### 200

A persistent stream of events

Content type: `text/event-stream`

Type: `object`. One of:

- Option 1 (`object`, schema `note_HeartBeatWorkspaceEvent`)
  - `event` (`string`, required): Possible values: `HeartBeatWorkspaceEvent`. Example: `HeartBeatWorkspaceEvent`.
  - `timestamp` (`string<date-time>`, required): Example: `2024-01-01T00:00:00Z`.
- Option 2 (`object`, schema `note_WorkspaceBillingChangedEvent`)
  - `id` (`string<uuid>`, required): Example: `dd57d134-d23b-457f-88fc-a4d7d6b284e5`.
  - `event` (`string`, required): Possible values: `WorkspaceBillingChangedEvent`. Example: `WorkspaceBillingChangedEvent`.
  - `data` (`object`, required): Refetch /billing/workspaces/{id}/entitlements.
    - `workspaceId` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.
    - `plan` (`string`, required): Example: `plus`.
- Option 3 (`object`, schema `note_WorkspaceDeletedEvent`)
  - `id` (`string<uuid>`, required): Example: `dd57d134-d23b-457f-88fc-a4d7d6b284e5`.
  - `event` (`string`, required): Possible values: `WorkspaceDeletedEvent`. Example: `WorkspaceDeletedEvent`.
  - `data` (`object`, required)
    - `id` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.
- Option 4 (`object`, schema `note_WorkspaceItemsUpdatedEvent`)
  - `id` (`string<uuid>`, required): Example: `dd57d134-d23b-457f-88fc-a4d7d6b284e5`.
  - `event` (`string`, required): Possible values: `WorkspaceItemsUpdatedEvent`. Example: `WorkspaceItemsUpdatedEvent`.
  - `data` (`object`, required)
    - `workspaceId` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.
- Option 5 (`object`, schema `note_WorkspaceMembersUpdatedEvent`)
  - `id` (`string<uuid>`, required): Example: `dd57d134-d23b-457f-88fc-a4d7d6b284e5`.
  - `event` (`string`, required): Possible values: `WorkspaceMembersUpdatedEvent`. Example: `WorkspaceMembersUpdatedEvent`.
  - `data` (`object`, required)
    - `workspaceId` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.
- Option 6 (`object`, schema `note_WorkspaceRenamedEvent`)
  - `id` (`string<uuid>`, required): Example: `dd57d134-d23b-457f-88fc-a4d7d6b284e5`.
  - `event` (`string`, required): Possible values: `WorkspaceRenamedEvent`. Example: `WorkspaceRenamedEvent`.
  - `data` (`object`, required)
    - `id` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.
    - `name` (`string`, schema `note_name-3`, required): Constraints: minLength: 1, maxLength: 255. Example: `Notopia`.
- Option 7 (`object`, schema `note_WorkspaceSlugChangedEvent`)
  - `id` (`string<uuid>`, required): Example: `dd57d134-d23b-457f-88fc-a4d7d6b284e5`.
  - `event` (`string`, required): Possible values: `WorkspaceSlugChangedEvent`. Example: `WorkspaceSlugChangedEvent`.
  - `data` (`object`, required)
    - `id` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.
    - `slug` (`string`, schema `note_slug`, required): Example: `notopia`.

Example:

```json
{
  "event": "HeartBeatWorkspaceEvent",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 400

Bad Request Error response

Content type: `application/json`

Type: `object` (schema `note_Error`).

- `code` (`string`, required): Error code. Example: `EXAMPLE_CODE`.
- `message` (`string`, required): Human-readable error message. Example: `Example error`.
- `more_info` (`string`): URL with more information about the error. Example: `https://api.notopia.com/docs/errors/EXAMPLE_CODE`.

Example:

```json
{
  "code": "BAD_REQUEST",
  "message": "The request could not be understood by the server due to malformed syntax",
  "more_info": "https://api.notopia.com/docs/errors/BAD_REQUEST"
}
```

### 401

Unauthorized Error response

Content type: `application/json`

Type: `object`.

Example:

```json
{}
```

### 500

Internal Server Error response

Content type: `application/json`

Type: `object` (see schema `note_Error` above).

Example:

```json
{
  "code": "INTERNAL_SERVER_ERROR",
  "message": "An internal server error occurred. Please try again later",
  "more_info": "https://api.notopia.com/docs/errors/INTERNAL_SERVER_ERROR"
}
```

## Code samples

**cURL**

```bash
curl http://api.notopia.localhost/note/workspaces/00000000-0000-4000-8000-000000000110/events \
  --header 'Authorization: Bearer note_Oauth2'
```
