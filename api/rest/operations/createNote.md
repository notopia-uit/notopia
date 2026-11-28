# Create note

`POST /note/notes`

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

Type: `object` (schema `note_CreateNoteRequest`).

- `name` (`string`, schema `note_name-2`, required): Can be empty string when creating but will be set to "Untitled Note" internally. Constraints: minLength: 1, maxLength: 255. Example: `PowerUp.ps1`.
- `icon` (`string | null`, schema `note_icon-2`, required): Example: `📓`.
- `folderId` (`string<uuid>`, schema `note_folderId`, required): Example: `123e4567-e89b-12d3-a456-426614174000`.

Example:

```json
{
  "name": "PowerUp.ps1",
  "icon": "📓",
  "folderId": "123e4567-e89b-12d3-a456-426614174000"
}
```

## Responses

### 201

Note successfully created

Headers:

| Name | Type | Description |
| --- | --- | --- |
| `Content-Location` | `string<uri>` | Example: `http://api.notopia.localhost/notes/note/000c2575-1847-5293-8117-2415a2328ef8`. |

Content type: `application/json`

Type: `object`.

- `id` (`string<uuid>`, schema `note_id-2`, required): Example: `000c2575-1847-5293-8117-2415a2328ef8`.

Example:

```json
{
  "id": "000c2575-1847-5293-8117-2415a2328ef8"
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
curl http://api.notopia.localhost/note/notes \
  --request POST \
  --header 'Authorization: Bearer note_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '{
  "name": "PowerUp.ps1",
  "icon": "📓",
  "folderId": "123e4567-e89b-12d3-a456-426614174000"
}'
```
