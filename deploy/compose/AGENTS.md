### Infrastructure

Layout:

- `infrastructure/` — Authentik, Traefik, Redpanda, Meilisearch, RustFS, dbgate
- `services/` — per-service dependencies (e.g. `note.yaml` = `note_postgresql` + `note_redis`)
- `dev/`, `monitoring/`
- `compose.yaml` — includes the above

#### Authentik

- Users/Client/Role list in `./deploy/compose/infrastructure/authentik-blueprints/users.yaml`
  > Read this file for login users. Must use Oauth2 flow, doesn't support direct access grant.
  > If you need debugging, use curl or anything related.
- Frontend App in `./deploy/compose/infrastructure/authentik-blueprints/app-web.yaml`
  (includes JWKS, callback URL, etc.)
- When editing a blueprint, authentik will reconcile itself

#### Traefik

- Config in `./deploy/compose/infrastructure/traefik/`: static (`traefik.yaml`) and
  dynamic (`dynamic/*.yaml`, one file per service)
- Dynamic config includes `agilezebra/jwt-middleware` for authentication; it authorizes
  using roles mapped from the OIDC `roles` claim, and injects user headers downstream
- Traefik mounts the host network, so it can reach services at `localhost:{port}`
- Backend services are routed by **path prefix on a single API host**, and each service
  excludes its own health path from the JWT middleware. For example:
  - `api.notopia.localhost/note`          -> note service (`/note/health` is unauthenticated)
  - `api.notopia.localhost/authorization` -> authorization service
  - `api.notopia.localhost/document`      -> document service
    (`/document/ws/document` is the Hocuspocus websocket route)
  - `api.notopia.localhost` (root)        -> api-web (Scalar spec viewer)
  - When adding a service, add `dynamic/{service}.yaml` and mirror this pattern.
    Any endpoint that must skip auth (e.g. a payment webhook) needs its own router
    without the JWT middleware.

#### Local dev hosts

All under `notopia.localhost`:

| Host | What |
| --- | --- |
| `web.notopia.localhost` | Next.js web app |
| `api.notopia.localhost` | All backend APIs, routed by path prefix (see above) |
| `authentik.notopia.localhost` | Authentik (IdP) |
| `traefik.notopia.localhost` | Traefik dashboard |
| `rustfs-api.notopia.localhost` | RustFS S3 API endpoint |
| `rustfs.notopia.localhost` | RustFS console |
| `meilisearch.notopia.localhost` | Meilisearch |
| `redpanda.notopia.localhost` | Redpanda console |
| `dbgate.notopia.localhost` | dbgate (DB browser) |
