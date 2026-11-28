# Stripe webhook endpoint

`POST /billing/webhooks/stripe`

## Servers

- `http://api.notopia.localhost`

## Parameters

### Header parameters

| Name | Type | Required | Description |
| --- | --- | --- | --- |
| `Stripe-Signature` | `string` | Yes | Example: `t=1700000000`. |

## Request body

Required.

Content type: `application/json`

Type: `object`.

- `additionalProperties` (`string`, additional properties)

Example (Stripe subscription updated event):

```json
{
  "id": "evt_test_example",
  "type": "customer.subscription.updated",
  "data": {
    "object": {
      "id": "sub_test_example"
    }
  }
}
```

## Responses

### 204

Event accepted

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
curl http://api.notopia.localhost/billing/webhooks/stripe \
  --request POST \
  --header 'Stripe-Signature: t=1700000000' \
  --header 'Content-Type: application/json' \
  --data '{
  "additionalProperties": "string"
}'
```
