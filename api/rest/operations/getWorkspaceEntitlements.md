# Get workspace entitlements

`GET /billing/workspaces/{workspaceId}/entitlements`

Available to workspace members. Returns the entitlements granted to the workspace by its current plan.

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
| `workspaceId` | `string<uuid>` | Yes | Example: `00000000-0000-4000-8000-000000000110`. |

## Responses

### 200

Entitlements of the workspace

Content type: `application/json`

Type: `object` (schema `billing_Entitlements`).

- `workspaceId` (`string<uuid>`, required): Example: `00000000-0000-4000-8000-000000000110`.
- `plan` (`string`, schema `billing_PlanTier`, required): Possible values: `free`, `plus`, `pro`. Example: `plus`.
- `status` (`string`, schema `billing_SubscriptionStatus`, required): `none` when the workspace never subscribed. Possible values: `none`, `incomplete`, `active`, `past_due`, `unpaid`, `canceled`, `incomplete_expired`. Example: `active`.
- `accessUntil` (`string<date-time> | null`, required): Example: `2026-12-22T08:50:14Z`.
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

Example (Free workspace without a subscription):

```json
{
  "accessUntil": null,
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
  "plan": "free",
  "status": "none",
  "workspaceId": "00000000-0000-4000-8000-000000000110"
}
```

Example (Plus workspace with an active subscription):

```json
{
  "accessUntil": "2026-12-22T15:50:14+07:00",
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
  "plan": "plus",
  "status": "active",
  "workspaceId": "00000000-0000-4000-8000-000000000110"
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
curl http://api.notopia.localhost/billing/workspaces/00000000-0000-4000-8000-000000000110/entitlements \
  --header 'Authorization: Bearer billing_Oauth2'
```
