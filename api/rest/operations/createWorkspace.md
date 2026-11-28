# Create workspace

`POST /note/workspaces`

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

## Request body

Required.

Content type: `application/json`

Type: `object` (schema `note_CreateWorkspaceRequest`).

- `name` (`string`, schema `note_name-3`, required): Constraints: minLength: 1, maxLength: 255. Example: `Notopia`.
- `slug` (`string`, schema `note_slug`, required): Example: `notopia`.

Example:

```json
{
  "name": "Notopia",
  "slug": "notopia"
}
```

## Responses

### 201

Workspace successfully created

Headers:

| Name | Type | Description |
| --- | --- | --- |
| `Content-Location` | `string<uri>` | Example: `http://api.notopia.localhost/workspaces/00000000-0000-4000-8000-000000000110`. |

Content type: `application/json`

Type: `object`.

- `id` (`string<uuid>`, schema `note_id-3`, required): Example: `00000000-0000-4000-8000-000000000110`.

Example:

```json
{
  "id": "00000000-0000-4000-8000-000000000110"
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
curl http://api.notopia.localhost/note/workspaces \
  --request POST \
  --header 'Authorization: Bearer note_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '{
  "name": "Notopia",
  "slug": "notopia"
}'
```
