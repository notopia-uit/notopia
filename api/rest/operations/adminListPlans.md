# List plans

`GET /billing/admin/plans`

Requires billing administrator access. Returns plan features, limits, and active prices.

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

## Responses

### 200

Plans

Content type: `application/json`

Type: `array<object>`.

- `tier` (`string`, schema `billing_PlanTier`, required): Possible values: `free`, `plus`, `pro`. Example: `plus`.
- `features` (`object`, schema `billing_Features`, required): Feature flags available to a plan. Example: `{"auto_commit":true,"workspace_graph":true}`.
  - `auto_commit` (`boolean`): Example: `true`.
  - `workspace_graph` (`boolean`): Example: `true`.
- `limits` (`object`, schema `billing_Limits`, required): Limits by key. -1 means unlimited. Example: `{"members":10,"published_notes":-1,"note_graph_max_depth":-1,"auto_commit_min_interval_seconds":300,"revision_history_days":30,"trash_retention_days":30,"ai_credits_monthly":200}`.
  - `members` (`integer<int64>`): Constraints: minimum: -1. Example: `10`.
  - `published_notes` (`integer<int64>`): Constraints: minimum: -1. Example: `-1`.
  - `note_graph_max_depth` (`integer<int64>`): Constraints: minimum: -1. Example: `-1`.
  - `auto_commit_min_interval_seconds` (`integer<int64>`): Constraints: minimum: -1. Example: `300`.
  - `revision_history_days` (`integer<int64>`): Constraints: minimum: -1. Example: `30`.
  - `trash_retention_days` (`integer<int64>`): Constraints: minimum: -1. Example: `30`.
  - `ai_credits_monthly` (`integer<int64>`): Constraints: minimum: -1. Example: `200`.
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

Example (Illustrative plan catalog (prices and IDs vary by environment)):

```json
[
  {
    "tier": "free",
    "features": {
      "auto_commit": false,
      "workspace_graph": false
    },
    "limits": {
      "ai_credits_monthly": 0,
      "auto_commit_min_interval_seconds": 0,
      "members": 3,
      "note_graph_max_depth": 5,
      "published_notes": 3,
      "revision_history_days": 7,
      "trash_retention_days": 7
    },
    "prices": []
  },
  {
    "tier": "plus",
    "features": {
      "auto_commit": true,
      "workspace_graph": true
    },
    "limits": {
      "ai_credits_monthly": 200,
      "auto_commit_min_interval_seconds": 300,
      "members": 10,
      "note_graph_max_depth": -1,
      "published_notes": -1,
      "revision_history_days": 30,
      "trash_retention_days": 30
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
      },
      {
        "id": "00000000-0000-4000-8000-000000000002",
        "plan": "plus",
        "interval": "year",
        "amount": 470400,
        "currency": "vnd",
        "lookupKey": "plus_year",
        "active": true,
        "createdAt": "2026-11-21T21:56:03.893201+07:00",
        "archivedAt": null
      }
    ]
  },
  {
    "tier": "pro",
    "features": {
      "auto_commit": true,
      "workspace_graph": true
    },
    "limits": {
      "ai_credits_monthly": 2000,
      "auto_commit_min_interval_seconds": 60,
      "members": -1,
      "note_graph_max_depth": -1,
      "published_notes": -1,
      "revision_history_days": -1,
      "trash_retention_days": 90
    },
    "prices": [
      {
        "id": "00000000-0000-4000-8000-000000000003",
        "plan": "pro",
        "interval": "month",
        "amount": 99000,
        "currency": "vnd",
        "lookupKey": "pro_month",
        "active": true,
        "createdAt": "2026-11-21T21:56:04.786267+07:00",
        "archivedAt": null
      },
      {
        "id": "00000000-0000-4000-8000-000000000004",
        "plan": "pro",
        "interval": "year",
        "amount": 950400,
        "currency": "vnd",
        "lookupKey": "pro_year",
        "active": true,
        "createdAt": "2026-11-21T21:56:05.619059+07:00",
        "archivedAt": null
      },
      {
        "id": "00000000-0000-4000-8000-000000000005",
        "plan": "pro",
        "interval": "lifetime",
        "amount": 2490000,
        "currency": "vnd",
        "lookupKey": "pro_lifetime",
        "active": true,
        "createdAt": "2026-11-21T21:56:06.448016+07:00",
        "archivedAt": null
      }
    ]
  }
]
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
curl http://api.notopia.localhost/billing/admin/plans \
  --header 'Authorization: Bearer billing_Oauth2'
```
