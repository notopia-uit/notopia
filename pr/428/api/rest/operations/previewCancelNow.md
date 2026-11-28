# Preview immediate cancellation

`GET /billing/workspaces/{workspaceId}/subscription/cancel-now/preview`

Requires workspace billing-management permission. Returns the refund quote for cancelling the current subscription immediately.

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

Refund quote

Content type: `application/json`

Type: `object` (schema `billing_RefundQuote`).

- `amount` (`integer<int64>`, required): Example: `29500`.
- `currency` (`string`, required): Example: `vnd`.
- `unusedSeconds` (`integer<int64>`, required): Example: `1296000`.
- `periodSeconds` (`integer<int64>`, required): Example: `2592000`.
- `payments` (`array<object>`, required): Refund breakdown by eligible paid invoice payment whose invoice coverage overlaps the current subscription period. A payment's coverage may differ from the current period, for example after a plan change. Example: `[{"paidAmount":49000,"amount":24500,"coverageSeconds":2592000,"unusedSeconds":1296000},{"paidAmount":10000,"amount":5000,"coverageSeconds":1296000,"unusedSeconds":648000}]`.
  - `paidAmount` (`integer<int64>`, required): Amount paid for this payment, in the quote currency. Example: `49000`.
  - `amount` (`integer<int64>`, required): Refund for this payment. Example: `24500`.
  - `coverageSeconds` (`integer<int64>`, required): Length of the payment's invoice coverage in seconds. Example: `2592000`.
  - `unusedSeconds` (`integer<int64>`, required): Unused portion of the payment's invoice coverage in seconds. Example: `1296000`.

Example (Prorated refund quote):

```json
{
  "amount": 24500,
  "currency": "vnd",
  "unusedSeconds": 1296000,
  "periodSeconds": 2592000,
  "payments": [
    {
      "paidAmount": 49000,
      "amount": 24500,
      "coverageSeconds": 2592000,
      "unusedSeconds": 1296000
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
curl http://api.notopia.localhost/billing/workspaces/00000000-0000-4000-8000-000000000110/subscription/cancel-now/preview \
  --header 'Authorization: Bearer billing_Oauth2'
```
