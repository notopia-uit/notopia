### Infrastructure

Layout:

- `infrastructure/` — Authentik, Traefik, Redpanda, Meilisearch, RustFS, dbgate
- `services/` — per-service dependencies (e.g. `note.yaml` = `note_postgresql` + `note_redis`)
- `dev/`, `monitoring/`
- `compose.yaml` — includes above

#### Authentik

- Users/Client/Role list in `./deploy/compose/infrastructure/authentik-blueprints/users.yaml`
  > Read file for login users. Must use Oauth2 flow; no direct access grant.
  > For debugging, use curl or similar.
- Frontend App in `./deploy/compose/infrastructure/authentik-blueprints/app-web.yaml`
  (includes JWKS, callback URL, etc.)
- Blueprint docs: <https://github.com/goauthentik/authentik/tree/main/website/docs/customize/blueprints>
- Authentik reconciles itself after blueprint edits

#### Traefik

- Config in `./deploy/compose/infrastructure/traefik/`: static (`traefik.yaml`) and
  dynamic (`dynamic/*.yaml`, one file per service)
- Dynamic config uses `agilezebra/jwt-middleware` for authentication. `notopia`
  OAuth scope emits `groups_json`, mapped to `X-Forwarded-Groups`; missing mapped claims
  remove client-supplied headers. Both JWT chains strip `X-Forwarded-Roles`.
  `jwt-admin` requires `notopia-admins` group.
- `node --test deploy/compose/infrastructure/traefik/gateway-security.test.mjs` tests
  disposable gateway with header-echo backend on `/billing/test-headers` and
  `/billing/admin/test-headers`. Set `GATEWAY_TEST_URL`, `GATEWAY_TEST_USER_TOKEN`,
  `GATEWAY_TEST_ADMIN_TOKEN`, `GATEWAY_TEST_NO_GROUPS_TOKEN`, and
  `GATEWAY_TEST_USER_GROUPS` (JSON array), plus `GATEWAY_TEST_API_WEB_TOKEN` and
  `GATEWAY_TEST_API_WEB_GROUPS` (nonempty JSON array of its actual groups). Never
  direct echo backend at production.
- Traefik mounts host network; can reach services at `localhost:{port}`
- Backend services route by **path prefix on a single API host**; each service
  excludes its health path from JWT middleware. Example:
  - `api.notopia.localhost/note`          -> note service (`/note/health` is unauthenticated)
  - `api.notopia.localhost/authorization` -> authorization service
  - `api.notopia.localhost/document`      -> document service
    (`/document/ws/document` is the Hocuspocus websocket route)
  - `api.notopia.localhost` (root)        -> api-web (Scalar spec viewer)
  - For new service, add `dynamic/{service}.yaml` and mirror pattern.
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

#### AI infrastructure

AI uses service digit **6**: HTTP and health share localhost:8086;
`services/ai.yaml` provides persistent `ai_redis` on localhost:6376 (AOF, noeviction).
`apps/ai/.env` points to it. `infrastructure/ollama.yaml` provides optional CPU-only
`ai_ollama` on localhost:11434, enabled by the `ai-local-llm` profile; models use a named
volume. Do not enable it when a host model server already occupies that port.
See the [AI service documentation](../../docs/src/services/ai/index.md) for startup,
model provider configuration and verification commands.
