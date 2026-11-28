# Get revisions

`GET /document/revisions`

Revisions older than the plan `revision_history_days` are purged daily; the latest is always kept. Filter with `trigger`.

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

### Query parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `documentId` | `string<uuid>` | Yes | Example: `000c2575-1847-5293-8117-2415a2328ef8`. |
| `page` | `integer` | No | Page number for pagination. Default: `1`. Constraints: minimum: 1. Example: `1`. |
| `limit` | `integer` | No | Number of items per page. Default: `20`. Constraints: minimum: 1, maximum: 100. Example: `20`. |
| `trigger` | `string` | No | Only return revisions created by this trigger. Possible values: `manual`, `auto`, `sweeper`. Example: `manual`. |

## Responses

### 200

Revisions

Content type: `application/json`

Type: `object`.

- `data` (`array<object>`, required)
  - `id` (`string<uuid>`, required): Example: `000c2575-1847-5293-8117-2415a2328ef8`.
  - `name` (`string | null`, required): Constraints: minLength: 1, maxLength: 255. Example: `Initial Draft`.
  - `createdAt` (`string<date-time>`, required): Example: `2024-01-01T12:00:00Z`.
  - `authorId` (`string | null`, required): Who the revision is attributed to. Null for revisions created before authors were recorded. Example: `42`.
  - `trigger` (`string`, schema `document_RevisionTrigger`, required): Possible values: `manual`, `auto`, `sweeper`. Example: `manual`.
  - `contributorIds` (`array<string>`, required): Users who edited the document since the previous revision, most recent last. Example: `["41","42"]`.
- `pagination` (`object`, schema `document_Pagination`, required)
  - `page` (`integer`, required): Current page number, starting from 1. Constraints: minimum: 1. Example: `1`.
  - `currentTotal` (`integer`, required): Number of items in the current page. Constraints: minimum: 0. Example: `20`.
  - `total` (`integer`, required): Total items across all pages. Constraints: minimum: 0. Example: `100`.
  - `totalPages` (`integer`, required): Total number of pages. Constraints: minimum: 0. Example: `5`.
  - `hasNext` (`boolean`, required): Whether there is a next page. Example: `true`.
  - `hasPrev` (`boolean`, required): Whether there is a previous page. Example: `false`.

Example:

```json
{
  "data": [
    {
      "id": "000c2575-1847-5293-8117-2415a2328ef8",
      "name": "Initial Draft",
      "createdAt": "2024-01-01T12:00:00Z",
      "authorId": "42",
      "trigger": "manual",
      "contributorIds": [
        "41",
        "42"
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "currentTotal": 20,
    "total": 100,
    "totalPages": 5,
    "hasNext": true,
    "hasPrev": false
  }
}
```

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
curl 'http://api.notopia.localhost/document/revisions?documentId=000c2575-1847-5293-8117-2415a2328ef8&page=1&limit=20&trigger=manual' \
  --header 'Authorization: Bearer document_Oauth2'
```
