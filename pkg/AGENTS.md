# pkg/

Cross-service Go code. Importable by any service; must not import `internal/`.

| Package | Contents |
| --- | --- |
| `common/config` | Reusable config structs (`Server`, `Sql`, `Redis`, `Log`, `Event`, `Authentik`, `Meilisearch`, `Service`, `General`). Services embed these |
| `common/handler` | `Cmd[C]` / `Query[Q,R]` interfaces + `Log*` and `Trace*` decorators |
| `common/http` | `NewGin` (recovery + otel + slog middleware), `GatewayUserAuth`, `SSEWrapper` |
| `metadata` | `ServiceName` / `ServiceVersion` named types — injected via Wire, used as otel resource attrs |
| `logging` | slog setup, plus Watermill and stdout adapters |
| `otel` | Tracer/meter/logger providers, propagator, resource, Watermill-Kafka instrumentation |
| `casbin` | Casbin → slog logger adapter |
| `api/` | **Generated** by oapi-codegen — `note/` (gin strict server + models), `share/` |
| `pb/` | **Generated** by buf — grpc + protovalidate stubs |

`api/` and `pb/` are generated: edit `api/` or `proto/` at the repo root and run
`nx gen api` / `nx gen proto`. Never edit `*.gen.go` or `*.pb.go`.

## Config structs

Tags are `mapstructure` + `yaml` + `json` + `default` + `validate` on every field. A missing
`default:"-"` or `validate` tag is usually a bug — the loader validates at startup, so a
service should fail fast rather than run with a zero value.

## Handler decorators

`Cmd`/`Query` are the CQRS contracts that `internal/*/app` implements. `NewLogCmd` and
`NewTraceCmd` wrap a handler to get uniform logging and an otel span named
`Cmd.Handle.{handlerName}` — wired once per handler in the service's `app/wire.go`, so
handlers themselves contain no logging or tracing boilerplate.

## `GatewayUserAuth`

Reads `X-Forwarded-ID` / `X-Forwarded-Email` — headers **injected by Traefik's
jwt-middleware after it validates the Authentik JWT**. It does not verify anything itself,
so it is only safe behind the gateway; never expose a service port directly. Retrieve with
`commonhttp.UserFromContext(ctx)` (requires gin's `ContextWithFallback`, already set in
`NewGin`).
