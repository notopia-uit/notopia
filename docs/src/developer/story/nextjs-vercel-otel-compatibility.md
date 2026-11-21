---
authorship:
  - ai
---

# Next.js OpenTelemetry compatibility

The web app uses `@vercel/otel` for Next.js instrumentation, including its Edge-compatible
registration path. The NestJS services use `@opentelemetry/sdk-node` independently. They run
in separate processes and do not need the same SDK generation.

## Why web has its own catalog

At the time of this change, `@vercel/otel` 2.3.0 declares peer dependencies on OpenTelemetry
SDK 2.x and experimental packages 0.2xx. SDK 2.x requires `@opentelemetry/api` below 1.10;
SDK 3.x requires API 1.10 or newer. A shared catalog would resolve incompatible versions for
one of the apps. The `web-otel` catalog in `pnpm-workspace.yaml` pins web to API 1.9.1,
SDK 2.12.1 and experimental packages 0.223.2. The default catalog tracks SDK 3.x for NestJS.

Web declares the `@vercel/otel` peer packages explicitly so pnpm resolves them as a matching
set. `@opentelemetry/sdk-logs` 0.220 and newer requires
`new BatchLogRecordProcessor({ exporter })`, not the former exporter-only constructor.

`@vercel/otel` 2.3.0 treats the literal `OTEL_SDK_DISABLED=false` as disabled. The web
bootstrap temporarily removes that value during synchronous registration and restores it
afterwards. `OTEL_SDK_DISABLED=true` still disables the SDK.

## Transport and homelab

Web uses OTLP over HTTP/protobuf for both traces and logs. `@vercel/otel` supports HTTP
trace exporters, but not its own gRPC trace exporter. Web's Node.js log exporter uses
`@opentelemetry/exporter-logs-otlp-proto`. The bootstrap rejects unsupported protocols
instead of silently sending to the wrong port. It also allows `http/json` for traces when
logs are independently configured as `http/protobuf`.

The web chart defaults to `otel.protocol: http/protobuf`. Its endpoint must point to an
OTLP HTTP receiver, not a gRPC-only endpoint. The homelab web values use
`http://alloy-alloy.monitoring.svc:4318`. Alloy's homelab `config.alloy` listens for OTLP
HTTP on 4318 and gRPC on 4317; its Helm values expose both ports. Other Notopia services
keep using gRPC on 4317. Alloy forwards traces to Tempo and logs to Loki. These are repository
settings, not proof of the live cluster's configuration.

## Upgrade when Vercel supports SDK 3.x

1. Check the published `@vercel/otel` peer dependencies and its changelog. Confirm support
   for API 1.10+, SDK 3.x, and matching experimental packages; do not infer compatibility
   from a successful install or a version number alone.
2. Update `@vercel/otel` and align web's complete OTel peer and exporter set with that
   supported generation. Remove the `web-otel` catalog only if the web and default catalog
   versions can actually be shared. Keep direct declarations needed for peer resolution.
3. Follow the [OTel JS 3.x migration guide](https://github.com/open-telemetry/opentelemetry-js/blob/main/doc/3.x/migration-guide.md):
   the logs API moves from `@opentelemetry/api-logs` into `@opentelemetry/api`, trace SDK
   packages change, and package export maps replace unsupported deep imports. Review
   Vercel's migration notes and the log processor constructor as well.
4. Check whether the `OTEL_SDK_DISABLED=false` workaround is still needed. Remove it only
   after verifying both `true` and `false` against the upgraded package.
5. Inspect the resolved web dependency graph and Next.js peer resolution for mixed or
   unsupported API/SDK versions. Build and lint web, and run its existing tests.
6. With telemetry enabled, send a trace and a log to a real or local OTLP HTTP receiver.
   Verify Next.js request spans, log export, context propagation, and the disabled-SDK path.
   Check the built container and the deployed collector separately; chart rendering alone
   does not establish connectivity or delivery.

See the [NestJS bundling story](./nestjs-bundling-otel.md) for backend instrumentation and
container constraints.
