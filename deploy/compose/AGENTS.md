### Infrastructure

#### Authentik

- Users/Client/Role list in `./deploy/compose/infrastructure/authentik-blueprints/users.yaml`
  > Read this files for login user. Must use Oauth2 flow, doesn't support direct access grant
  > If need debugging, use curl or any relate
- Frontend App in `./deploy/compose/infrastructure/authentik-blueprints/app-web.yaml` (include JWKS, callback URL, etc.)
- When editing blueprint, authentik will reconcile itself

#### Traefik

- Traefik config in `./deploy/compose/infrastructure/traefik/`, which include static and dynamic config
- Dynamic including `agilezebra/jwt-middleware` for authenticate
- Traefik will mount host network into itself, so it can access other services with `localhost:{port}`
- General local dev domain is `notopia.localhost`:
  - Web: `web.notopia.localhost`
  - API: `api.notopia.localhost`, include `api-web` for root, and `course`, `billing`, `blog` for each service
  - S3: `rustfs-api.notopia.localhost`

#### Authentik

- Users/Client/Role list in `./deploy/compose/infrastructure/authentik-blueprints/users.yaml`
  > Read this files for login user. Must use Oauth2 flow, doesn't support direct access grant
  > If need debugging, use curl or any relate
- Frontend App in `./deploy/compose/infrastructure/authentik-blueprints/app-web.yaml` (include JWKS, callback URL, etc.)
- When editing blueprint, authentik will reconcile itself

#### Traefik

- Traefik config in `./deploy/compose/infrastructure/traefik/`, which include static and dynamic config
- Dynamic including `agilezebra/jwt-middleware` for authenticate, authorize with role mapped from OIDC `roles` claim
- Traefik will mount host network into itself, so it can access other services with `localhost:{port}`
- General local dev domain is `notopia.localhost`:
  - Web: `web.notopia.localhost`
  - API: `api.notopia.localhost`
  - S3: `rustfs.notopia.localhost`
