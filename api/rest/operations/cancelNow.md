# Cancel subscription now

`POST /billing/workspaces/{workspaceId}/subscription/cancel-now`

Requires workspace billing-management permission. Cancels the current subscription immediately and requests a prorated refund when applicable.

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

Workspace billing state after cancel

Content type: `application/json`

Type: `object` (schema `billing_WorkspaceBilling`).

- `subscription` (`object`, required)
  - `id` (`string<uuid>`, required): Example: `00000000-0000-4000-8000-000000000201`.
  - `plan` (`string`, schema `billing_PlanTier`, required): Possible values: `free`, `plus`, `pro`. Example: `plus`.
  - `interval` (`string`, schema `billing_BillingInterval`, required): Possible values: `month`, `year`, `lifetime`. Example: `month`.
  - `status` (`string`, schema `billing_SubscriptionStatus`, required): `none` when the workspace never subscribed. Possible values: `none`, `incomplete`, `active`, `past_due`, `unpaid`, `canceled`, `incomplete_expired`. Example: `active`.
  - `currentPeriodStart` (`string<date-time> | null`, required): Example: `2026-11-22T08:50:14Z`.
  - `currentPeriodEnd` (`string<date-time> | null`, required): Example: `2026-12-22T08:50:14Z`.
  - `cancelAtPeriodEnd` (`boolean`, required): Example: `false`.
  - `canceledAt` (`string<date-time> | null`, required)
  - `accessUntil` (`string<date-time> | null`, required): Example: `2026-12-22T08:50:14Z`.
  - `refunds` (`array<object>`, required): Example: `[]`.
    - `id` (`string<uuid>`, required): Example: `00000000-0000-4000-8000-000000000301`.
    - `amount` (`integer<int64>`, required): Example: `24500`.
    - `currency` (`string`, required): Example: `vnd`.
    - `reason` (`string`, required): Possible values: `cancel_now`, `lifetime_window`, `replaced_by_lifetime`, `workspace_deleted`, `duplicate_checkout`, `external`. Example: `cancel_now`.
    - `status` (`string`, required): Possible values: `pending`, `succeeded`, `failed`, `canceled`. Example: `succeeded`.
    - `createdAt` (`string<date-time>`, required): Example: `2026-10-11T00:00:00Z`.
- `payer` (`object`, required)
  - `id` (`string`, required): Example: `112`.
  - `email` (`string`, required): Example: `trannguyenthaibinh46@gmail.com`.
  - `name` (`string`, required): Example: `Kevin Nitro`.
  - `missingSince` (`string<date-time> | null`, required)

Example (Canceled subscription with a pending refund):

```json
{
  "payer": {
    "id": "112",
    "email": "trannguyenthaibinh46@gmail.com",
    "name": "Kevin Nitro",
    "missingSince": null
  },
  "subscription": {
    "id": "00000000-0000-4000-8000-000000000201",
    "plan": "plus",
    "interval": "month",
    "status": "canceled",
    "currentPeriodStart": "2026-11-22T08:50:14Z",
    "currentPeriodEnd": "2026-12-22T08:50:14Z",
    "cancelAtPeriodEnd": false,
    "canceledAt": "2026-12-07T08:50:14Z",
    "accessUntil": null,
    "refunds": [
      {
        "id": "00000000-0000-4000-8000-000000000301",
        "amount": 24500,
        "currency": "vnd",
        "reason": "cancel_now",
        "status": "pending",
        "createdAt": "2026-12-07T08:50:14Z"
      }
    ]
  }
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

### 409

Conflict with current billing state

Content type: `application/json`

Type: `object` (see schema `billing_Error` above).

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
curl http://api.notopia.localhost/billing/workspaces/00000000-0000-4000-8000-000000000110/subscription/cancel-now \
  --request POST \
  --header 'Authorization: Bearer billing_Oauth2'
```
