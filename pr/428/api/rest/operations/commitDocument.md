# Commit document

`POST /document/documents/{documentId}/commit`

Creates a revision of the current content. Requires write permission. Plus/pro workspaces also auto-commit (`auto_commit`), so this is only needed to force a revision earlier.

## Servers

- `http://api.notopia.localhost`

## Authorization

Use one of the following options:

- Option 1:
  - `document_Oauth2`: OAuth 2.0. Required scopes: `openid`. OAuth2 Security Scheme
    - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
    - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/
- Option 2:
  - `document_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). Required scopes: `openid`. OpenID Connect Security Scheme

## Parameters

### Path parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `documentId` | `string<uuid>` | Yes | Unique identifier of the document (note). Example: `000c2575-1847-5293-8117-2415a2328ef8`. |

## Responses

### 200

Nothing changed since the last revision; returns the latest revision

Content type: `application/json`

Type: `object` (schema `document_CommittedRevision`).

- `id` (`string<uuid>`, schema `document_id-2`, required): Example: `000c2575-1847-5293-8117-2415a2328ef8`.

Example:

```json
{
  "id": "000c2575-1847-5293-8117-2415a2328ef8"
}
```

### 201

Document revision committed successfully

Content type: `application/json`

Type: `object` (see schema `document_CommittedRevision` above).

Example:

```json
{
  "id": "000c2575-1847-5293-8117-2415a2328ef8"
}
```

### 204

Nothing changed and the document has no revision yet

### 400

Bad Request Error response

Content type: `application/json`

Type: `object` (schema `document_Error`).

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

### 403

Forbidden Error response

Content type: `application/json`

Type: `object` (see schema `document_Error` above).

Example:

```json
{
  "code": "FORBIDDEN",
  "message": "You do not have permission to access this resource",
  "more_info": "https://api.notopia.com/docs/errors/FORBIDDEN"
}
```

### 404

Not Found Error response

Content type: `application/json`

Type: `object` (see schema `document_Error` above).

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

Type: `object` (see schema `document_Error` above).

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
curl http://api.notopia.localhost/document/documents/000c2575-1847-5293-8117-2415a2328ef8/commit \
  --request POST \
  --header 'Authorization: Bearer document_Oauth2'
```
