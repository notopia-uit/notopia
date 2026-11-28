# Update plan

`PATCH /billing/admin/plans/{tier}`

Requires billing administrator access. Updates the selected plan's features and limits.

## Servers

- `http://api.notopia.localhost`

## Authorization

Use one of the following options:

- Option 1:
  - `billing_Oauth2`: OAuth 2.0. Required scopes: `openid`. OAuth2 Security Scheme
    - Flow `authorizationCode`: authorization URL http://authentik.notopia.localhost/application/o/authorize/, token URL http://authentik.notopia.localhost/application/o/token/
    - Flow `password`: token URL http://authentik.notopia.localhost/application/o/token/
- Option 2:
  - `billing_OIDC`: OpenID Connect (http://authentik.notopia.localhost/application/o/api-web/.well-known/openid-configuration). Required scopes: `openid`. OpenID Connect Security Scheme

## Parameters

### Path parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `tier` | `string` | Yes | Possible values: `free`, `plus`, `pro`. Example: `plus`. |

## Request body

Required.

Content type: `application/json`

Type: `object` (schema `billing_UpdatePlanRequest`).

- `features` (`object`, schema `billing_Features`): Feature flags available to a plan. Example: `{"auto_commit":true,"workspace_graph":true}`.
  - `auto_commit` (`boolean`): Example: `true`.
  - `workspace_graph` (`boolean`): Example: `true`.
- `limits` (`object`, schema `billing_Limits`): Limits by key. -1 means unlimited. Example: `{"members":10,"published_notes":-1,"note_graph_max_depth":-1,"auto_commit_min_interval_seconds":300,"revision_history_days":30,"trash_retention_days":30,"ai_credits_monthly":200}`.
  - `members` (`integer<int64>`): Constraints: minimum: -1. Example: `10`.
  - `published_notes` (`integer<int64>`): Constraints: minimum: -1. Example: `-1`.
  - `note_graph_max_depth` (`integer<int64>`): Constraints: minimum: -1. Example: `-1`.
  - `auto_commit_min_interval_seconds` (`integer<int64>`): Constraints: minimum: -1. Example: `300`.
  - `revision_history_days` (`integer<int64>`): Constraints: minimum: -1. Example: `30`.
  - `trash_retention_days` (`integer<int64>`): Constraints: minimum: -1. Example: `30`.
  - `ai_credits_monthly` (`integer<int64>`): Constraints: minimum: -1. Example: `200`.

Example (Plus plan features and limits):

```json
{
  "features": {
    "auto_commit": true,
    "workspace_graph": true
  },
  "limits": {
    "members": 10,
    "published_notes": -1,
    "note_graph_max_depth": -1,
    "auto_commit_min_interval_seconds": 300,
    "revision_history_days": 30,
    "trash_retention_days": 30,
    "ai_credits_monthly": 200
  }
}
```

## Responses

### 200

Updated plan

Content type: `application/json`

Type: `object` (schema `billing_Plan`).

- `tier` (`string`, schema `billing_PlanTier`, required): Possible values: `free`, `plus`, `pro`. Example: `plus`.
- `features` (`object`, see schema `billing_Features`, required): Feature flags available to a plan. Example: `{"auto_commit":true,"workspace_graph":true}`.
- `limits` (`object`, see schema `billing_Limits`, required): Limits by key. -1 means unlimited. Example: `{"members":10,"published_notes":-1,"note_graph_max_depth":-1,"auto_commit_min_interval_seconds":300,"revision_history_days":30,"trash_retention_days":30,"ai_credits_monthly":200}`.
- `prices` (`array<object>`, required)
  - `id` (`string<uuid>`, required): Example: `00000000-0000-4000-8000-000000000001`.
  - `plan` (`string`, schema `billing_PlanTier`, required): Possible values: `free`, `plus`, `pro`. Example: `plus`.
  - `interval` (`string`, schema `billing_BillingInterval`, required): Possible values: `month`, `year`, `lifetime`. Example: `month`.
  - `amount` (`integer<int64>`, required): Example: `49000`.
  - `currency` (`string`, required): Example: `vnd`.
  - `lookupKey` (`string`, required): Example: `plus_month`.
  - `active` (`boolean`, required): Example: `true`.
  - `createdAt` (`string<date-time>`, required): Example: `2026-11-21T00:00:00Z`.
  - `archivedAt` (`string<date-time> | null`, required)

Example (Updated Plus plan):

```json
{
  "tier": "plus",
  "features": {
    "auto_commit": true,
    "workspace_graph": true
  },
  "limits": {
    "members": 10,
    "published_notes": -1,
    "note_graph_max_depth": -1,
    "auto_commit_min_interval_seconds": 300,
    "revision_history_days": 30,
    "trash_retention_days": 30,
    "ai_credits_monthly": 200
  },
  "prices": [
    {
      "id": "00000000-0000-4000-8000-000000000001",
      "plan": "plus",
      "interval": "month",
      "amount": 49000,
      "currency": "vnd",
      "lookupKey": "plus_month",
      "active": true,
      "createdAt": "2026-11-21T00:00:00Z",
      "archivedAt": null
    }
  ]
}
```

### 400

Bad Request Error response

Content type: `application/json`

Type: `object` (schema `billing_Error`).

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

Type: `object` (see schema `billing_Error` above).

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

Type: `object` (see schema `billing_Error` above).

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

Type: `object` (see schema `billing_Error` above).

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
curl http://api.notopia.localhost/billing/admin/plans/plus \
  --request PATCH \
  --header 'Authorization: Bearer billing_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '{
  "features": {
    "auto_commit": true,
    "workspace_graph": true
  },
  "limits": {
    "members": 10,
    "published_notes": -1,
    "note_graph_max_depth": -1,
    "auto_commit_min_interval_seconds": 300,
    "revision_history_days": 30,
    "trash_retention_days": 30,
    "ai_credits_monthly": 200
  }
}'
```
