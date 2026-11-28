# Get workspace tree

`GET /note/workspaces/{workspaceId}/tree`

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

### Query parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `rootFolderId` | `string<uuid>` | No | Example: `123e4567-e89b-12d3-a456-426614174000`. |
| `includeTrashed` | `boolean` | No | Default: `false`. |
| `depth` | `integer` | No | Constraints: minimum: 1. |
| `sort` | `object` | No |  |

## Responses

### 200

OK

Content type: `application/json`

Type: `object` (schema `note_WorkspaceTreeFolder`).

- `id` (`string<uuid>`, schema `note_id`, required): Example: `123e4567-e89b-12d3-a456-426614174000`.
- `name` (`string`, schema `note_name`, required): Constraints: minLength: 1, maxLength: 255. Example: `Work Notes`.
- `icon` (`string | null`, schema `note_icon`, required): Example: `📁`.
- `notes` (`array<object>`, required)
  - `id` (`string<uuid>`, schema `note_id-2`, required): Example: `000c2575-1847-5293-8117-2415a2328ef8`.
  - `name` (`string`, schema `note_name-2`, required): Can be empty string when creating but will be set to "Untitled Note" internally. Constraints: minLength: 1, maxLength: 255. Example: `PowerUp.ps1`.
  - `icon` (`string | null`, schema `note_icon-2`, required): Example: `📓`.
  - `updatedAt` (`string<date-time>`, schema `note_updatedAt`, required): Example: `2024-06-01T12:34:56Z`.
- `children` (`array<object>`, required)
- `updatedAt` (`string<date-time>`, schema `note_updatedAt-2`, required): Example: `2024-06-01T12:00:00Z`.

Example:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Work Notes",
  "icon": "📁",
  "notes": [
    {
      "id": "000c2575-1847-5293-8117-2415a2328ef8",
      "name": "PowerUp.ps1",
      "icon": "📓",
      "updatedAt": "2024-06-01T12:34:56Z"
    }
  ],
  "children": [
    {}
  ],
  "updatedAt": "2024-06-01T12:00:00Z"
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

### 404

Not Found Error response

Content type: `application/json`

Type: `object` (see schema `note_Error` above).

Example:

```json
{
  "code": "NOT_FOUND",
  "message": "The requested resource was not found",
  "more_info": "https://api.notopia.com/docs/errors/NOT_FOUND"
}
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
curl 'http://api.notopia.localhost/note/workspaces/00000000-0000-4000-8000-000000000110/tree?rootFolderId=123e4567-e89b-12d3-a456-426614174000' \
  --header 'Authorization: Bearer note_Oauth2'
```
