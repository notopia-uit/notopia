# Search users

`GET /note/search-users`

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

### Query parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `keyword` | `string` | Yes | Search keyword.<br>In authentik, it may search based on username, email, full name, but not user ID (PK).<br><br>Example: `user`. |
| `isActive` | `boolean` | No | If provided, the search will be limited to (in)active users |
| `limit` | `integer` | No | Number of items per page. Default: `20`. Constraints: minimum: 1, maximum: 100. Example: `20`. |
| `excludeMemberInWorkspaceId` | `string<uuid>` | No | If provided, the search will be exclude users who are members of the specified workspace.<br>Note that this will peform after getting users from identity provider, so it doesn't conform the `limit` parameter.<br><br>Example: `00000000-0000-4000-8000-000000000110`. |

## Responses

### 200

A list of workspace members

Content type: `application/json`

Type: `array<object>`.

- `id` (`string`, required): User ID from Authentik (need to change subject mode to User's ID instead of hashed). Example: `100`.
- `username` (`string`, required): Username from Authentik. Example: `user`.
- `name` (`string | null`, required): Full name from Authentik. Example: `User`.
- `email` (`string<email> | null`, required): Email from Authentik. Example: `user@notopia.com`.

Example:

```json
[
  {
    "id": "100",
    "username": "user",
    "name": "User",
    "email": "user@notopia.com"
  }
]
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

Type: `object` (schema `note_Error`).

- `code` (`string`, required): Error code. Example: `EXAMPLE_CODE`.
- `message` (`string`, required): Human-readable error message. Example: `Example error`.
- `more_info` (`string`): URL with more information about the error. Example: `https://api.notopia.com/docs/errors/EXAMPLE_CODE`.

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
curl 'http://api.notopia.localhost/note/search-users?keyword=user&limit=20&excludeMemberInWorkspaceId=00000000-0000-4000-8000-000000000110' \
  --header 'Authorization: Bearer note_Oauth2'
```
