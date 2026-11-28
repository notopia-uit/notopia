# Create price

`POST /billing/admin/prices`

Requires billing administrator access. Replaces the active price for the same plan, billing interval, and currency; the previous price is archived.

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

## Request body

Required.

Content type: `application/json`

Type: `object` (schema `billing_CreatePriceRequest`).

- `plan` (`string`, required): Possible values: `plus`, `pro`. Example: `plus`.
- `interval` (`string`, schema `billing_BillingInterval`, required): Possible values: `month`, `year`, `lifetime`. Example: `month`.
- `amount` (`integer<int64>`, required): Constraints: minimum: 1. Example: `49000`.
- `currency` (`string`, required): Example: `vnd`.

Example (New Plus monthly price):

```json
{
  "plan": "plus",
  "interval": "month",
  "amount": 49000,
  "currency": "vnd"
}
```

## Responses

### 201

Created price

Content type: `application/json`

Type: `object` (schema `billing_Price`).

- `id` (`string<uuid>`, required): Example: `00000000-0000-4000-8000-000000000001`.
- `plan` (`string`, schema `billing_PlanTier`, required): Possible values: `free`, `plus`, `pro`. Example: `plus`.
- `interval` (`string`, schema `billing_BillingInterval`, required): Possible values: `month`, `year`, `lifetime`. Example: `month`.
- `amount` (`integer<int64>`, required): Example: `49000`.
- `currency` (`string`, required): Example: `vnd`.
- `lookupKey` (`string`, required): Example: `plus_month`.
- `active` (`boolean`, required): Example: `true`.
- `createdAt` (`string<date-time>`, required): Example: `2026-11-21T00:00:00Z`.
- `archivedAt` (`string<date-time> | null`, required)

Example (Created Plus monthly price):

```json
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
curl http://api.notopia.localhost/billing/admin/prices \
  --request POST \
  --header 'Authorization: Bearer billing_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '{
  "plan": "plus",
  "interval": "month",
  "amount": 49000,
  "currency": "vnd"
}'
```
