# Update workspace members

`PUT /note/workspaces/{workspaceId}/members`

Plan limit `members` (free 3, plus 10, pro unlimited). Adding beyond the limit returns 402. The billing payer cannot lose the owner role while a recurring subscription is active (409 `billingPayerMustTransfer`).

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

## Request body

Required.

Content type: `application/json`

Type: `array<object>`.

- `id` (`string`, schema `note_id-4`, required): User ID from Authentik (need to change subject mode to User's ID instead of hashed). Example: `100`.
- `name` (`string | null`, schema `note_name-4`): Full name from Authentik. Example: `User`.
- `role` (`string`, schema `note_WorkspaceRole`, required): Possible values: `owner`, `editor`, `viewer`. Example: `owner`.

Example:

```json
[
  {
    "id": "100",
    "name": "User",
    "role": "owner"
  }
]
```

## Responses

### 204

Workspace members updated successfully

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

### 402

Feature or limit not included in the workspace plan

Content type: `application/json`

Error for a request blocked by the workspace plan. `code` is one of:
- `featureNotInPlan` (details.kind=feature): the feature is not included in the plan
- `planLimitExceeded` (details.kind=limit): the plan's usage limit has been reached

Type: `object` (schema `note_PaymentRequiredError`).

- `code` (`string`, required): Error code. Example: `EXAMPLE_CODE`.
- `message` (`string`, required): Human-readable error message. Example: `Example error`.
- `more_info` (`string`): URL with more information about the error. Example: `https://api.notopia.com/docs/errors/EXAMPLE_CODE`.
- `details` (`object`)
  - `kind` (`string`, required): Possible values: `feature`, `limit`. Example: `limit`.
  - `key` (`string`, required): Example: `members`.
  - `current` (`integer<int64>`): Example: `5`.
  - `max` (`integer<int64>`): Example: `5`.
  - `requiredPlan` (`string`, required): Possible values: `plus`, `pro`. Example: `plus`.

Example:

```json
{
  "code": "planLimitExceeded",
  "message": "Feature or limit not included in the workspace plan",
  "more_info": "https://api.notopia.com/docs/errors/planLimitExceeded",
  "details": {
    "kind": "limit",
    "key": "members",
    "current": 5,
    "max": 5,
    "requiredPlan": "plus"
  }
}
```

### 403

Forbidden Error response

Content type: `application/json`

Type: `object` (see schema `note_Error` above).

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

Type: `object` (see schema `note_Error` above).

Example:

```json
{
  "code": "NOT_FOUND",
  "message": "The requested resource was not found",
  "more_info": "https://api.notopia.com/docs/errors/NOT_FOUND"
}
```

### 409

The billing payer must transfer the owner role before it can be removed while a recurring subscription is active

Content type: `application/json`

Type: `object` (see schema `note_Error` above).

Example:

```json
{
  "code": "EXAMPLE_CODE",
  "message": "Example error",
  "more_info": "https://api.notopia.com/docs/errors/EXAMPLE_CODE"
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

### 503

Billing service unavailable

Content type: `application/json`

Type: `object` (see schema `note_Error` above).

Example:

```json
{
  "code": "SERVICE_UNAVAILABLE",
  "message": "Billing service unavailable. Please try again later",
  "more_info": "https://api.notopia.com/docs/errors/SERVICE_UNAVAILABLE"
}
```

## Code samples

**cURL**

```bash
curl http://api.notopia.localhost/note/workspaces/00000000-0000-4000-8000-000000000110/members \
  --request PUT \
  --header 'Authorization: Bearer note_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '[
  {
    "id": "100",
    "name": "User",
    "role": "owner"
  }
]'
```
