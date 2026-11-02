---
title: Auth and Gateway
order: 3
---

# Auth and Gateway

## Public traffic

All public traffic enters through Traefik, routed by path prefix on a single API host
(`/note`, `/authorization`, `/document`, `/billing`, ...). Traefik validates the Authentik JWT
with a JWT middleware, maps the OIDC `roles` claim, and injects user headers downstream. Each
service excludes its own health path from the middleware, and endpoints that must skip auth
(such as the Stripe webhook) have their own router.

- Go services read the user with `commonhttp.GatewayUserAuth()`.
- NestJS services use `UserGuard` and the `@User()` decorator.
- Exception: the `apps/document` websocket (`/document/ws/document`) validates the JWT itself.

## Internal traffic

Service-to-service calls use gRPC (definitions in `proto/`, see [gRPC reference](../api/grpc/)).
A service never calls another service's public REST API internally. For example, Note and
Billing call the Authorization service over gRPC for permission checks, and Note calls Billing
for entitlements on every check without caching.
