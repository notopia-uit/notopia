# pkg/

Cross-service Go code. Any service can import; must not import `internal/`.

| Package | Contents |
| --- | --- |
| `common/config` | Reusable config structs (`Server`, `Sql`, `Redis`, `Log`, `Event`, `Authentik`, `Meilisearch`, `Service`, `General`, `Identity`). Services embed them. `Identity` holds `AdminGroup` and `ServiceAccountUserID`, set by `NOTOPIA_{SERVICE}_IDENTITY_ADMIN_GROUP` / `..._SERVICE_ACCOUNT_USER_ID` |
| `common/handler` | `Cmd[C]` / `Query[Q,R]` interfaces + `Log*` and `Trace*` decorators |
| `common/http` | `NewGin` (recovery + otel + slog middleware), `GatewayUserAuth`, `RequireGroup`, `SSEWrapper` |
| `cron` | gocron v2 scheduler with Redis distributed lock (`NewRedisLocker`, `WithTries(1)`, 3s auto-extend). `NewScheduler(locker, logger, time.Local)`; never hard-code zone. `Register` validates `Job{Name, Schedule, Timeout, Enabled, Run}`, skips disabled jobs; build from `commonconfig.CronJob` with `JobFromConfig`. `Locker` injected (`NewRedisLocker`). Names service-prefixed (`note:...`). Register from `internal/{svc}/controller/cron/`. Jobs must be idempotent |
| `metadata` | `ServiceName` / `ServiceVersion` named types — injected via Wire, used as otel resource attrs |
| `logging` | slog setup, Watermill and stdout adapters |
| `otel` | Tracer/meter/logger providers, propagator, resource, Watermill-Kafka instrumentation |
| `casbin` | Casbin → slog logger adapter |
| `api/` | **Generated** by oapi-codegen — `note/` (gin strict server + models), `share/` |
| `pb/` | **Generated** by buf — grpc + protovalidate stubs |

`api/` and `pb/` generated: edit `api/` or `proto/` at repo root, run
`nx gen api` / `nx gen proto`. Never edit `*.gen.go` or `*.pb.go`.

## Config structs

Every field has `mapstructure` + `yaml` + `json` + `default` + `validate` tags. Missing
`default:"-"` or `validate` tag usually a bug — loader validates at startup;
service should fail fast, not run with zero value.

## Handler decorators

`Cmd`/`Query` are CQRS contracts implemented by `internal/*/app`. `NewLogCmd` and
`NewTraceCmd` wrap handler for uniform logging and otel span named
`Cmd.Handle.{handlerName}` — wired once per handler in service's `app/wire.go`, so
handlers need no logging or tracing boilerplate.

## `GatewayUserAuth`

Reads `X-Forwarded-ID` / `X-Forwarded-Email` / `X-Forwarded-Name`, plus `X-Forwarded-Groups`
and `X-Forwarded-Roles` (JSON arrays). Locally Traefik validates JWT, maps Authentik's `groups_json`
claim to `X-Forwarded-Groups`; Istio supplies gateway headers in cluster. Traefik strips client-supplied roles, doesn't map roles claim;
other gateways must also strip or set that header. Middleware verifies nothing itself,
so safe only behind gateway; never expose service port directly. Retrieve with
`commonhttp.UserFromContext(ctx)` (requires gin's `ContextWithFallback`, already set in
`NewGin`). `RequireGroup(group)` aborts with 403 `{code: "forbidden"}` if user lacks group.
