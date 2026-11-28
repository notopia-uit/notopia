# Create folder

`POST /note/folders`

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

Type: `object` (schema `note_CreateFolderRequest`).

- `name` (`string`, schema `note_name`, required): Constraints: minLength: 1, maxLength: 255. Example: `Work Notes`.
- `icon` (`string | null`, schema `note_icon`, required): Example: `📁`.
- `parentId` (`string<uuid> | null`, schema `note_parentId`, required): Example: `cc7a2e47-6267-4c60-8d36-257f849777e4`.
- `workspaceId` (`string<uuid>`, schema `note_workspaceId`, required): Example: `00000000-0000-4000-8000-000000000110`.

Example:

```json
{
  "name": "Work Notes",
  "icon": "📁",
  "parentId": "cc7a2e47-6267-4c60-8d36-257f849777e4",
  "workspaceId": "00000000-0000-4000-8000-000000000110"
}
```

## Responses

### 201

Folder successfully created

Headers:

| Name | Type | Description |
| --- | --- | --- |
| `Content-Location` | `string<uri>` | Example: `http://api.notopia.localhost/notes/folder/123e4567-e89b-12d3-a456-426614174000`. |

Content type: `application/json`

Type: `object`.

- `id` (`string<uuid>`, schema `note_id`, required): Example: `123e4567-e89b-12d3-a456-426614174000`.

Example:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000"
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
curl http://api.notopia.localhost/note/folders \
  --request POST \
  --header 'Authorization: Bearer note_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '{
  "name": "Work Notes",
  "icon": "📁",
  "parentId": "cc7a2e47-6267-4c60-8d36-257f849777e4",
  "workspaceId": "00000000-0000-4000-8000-000000000110"
}'
```
