# Get document attachment upload url

`GET /document/documents/{documentId}/attachment-url`

Get presigned URL for document attachment upload

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

### Query parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `filename` | `string` | Yes | Example: `sam_seg.mp4`. |

## Responses

### 200

Presigned URL

Content type: `application/json`

Type: `object`.

- `uploadUrl` (`string<uri>`, required): Example: `http://rustfs-api.notopia.localhost/document/upload-url`.
- `url` (`string<uri>`, required): Example: `http://rustfs-api.notopia.localhost/document/document-attachments/d4c6b66a-03a8-486d-a139-ab72635c009b/a64cb105-4924-4441-8611-bbc91bbaa1b9-sam_seg.mp4`.

Example:

```json
{
  "uploadUrl": "http://rustfs-api.notopia.localhost/document/upload-url",
  "url": "http://rustfs-api.notopia.localhost/document/document-attachments/d4c6b66a-03a8-486d-a139-ab72635c009b/a64cb105-4924-4441-8611-bbc91bbaa1b9-sam_seg.mp4"
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
curl 'http://api.notopia.localhost/document/documents/000c2575-1847-5293-8117-2415a2328ef8/attachment-url?filename=sam_seg.mp4' \
  --header 'Authorization: Bearer document_Oauth2'
```
