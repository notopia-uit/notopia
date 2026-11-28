# Create checkout session

`POST /billing/workspaces/{workspaceId}/checkout-session`

Requires workspace billing-management permission. Returns a URL to redirect the browser to Stripe Checkout to purchase the selected active price.

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

## Request body

Required.

Content type: `application/json`

Type: `object` (schema `billing_CheckoutSessionRequest`).

- `priceId` (`string<uuid>`, required): Example: `00000000-0000-4000-8000-000000000001`.

Example (Plus monthly checkout request):

```json
{
  "priceId": "00000000-0000-4000-8000-000000000001"
}
```

## Responses

### 200

Checkout session URL. Redirect the browser to this URL to complete the purchase.

Content type: `application/json`

Type: `object` (schema `billing_RedirectURL`).

- `url` (`string<uri>`, required): Example: `https://checkout.stripe.com/c/pay/cs_test_123`.

Example (Stripe Checkout session):

```json
{
  "url": "https://checkout.stripe.com/c/pay/cs_test_example"
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
curl http://api.notopia.localhost/billing/workspaces/00000000-0000-4000-8000-000000000110/checkout-session \
  --request POST \
  --header 'Authorization: Bearer billing_Oauth2' \
  --header 'Content-Type: application/json' \
  --data '{
  "priceId": "00000000-0000-4000-8000-000000000001"
}'
```
