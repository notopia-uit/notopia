# NestJS gRPC: no `.proto` files at runtime, but `@grpc/proto-loader` must be installed

> Related: [`nestjs-bundling-otel-2026-10-blocknote-055.md`](./nestjs-bundling-otel-2026-10-blocknote-055.md) §4.1.

## TL;DR

|                                  | Needed at runtime? | Why                                                                                                                     |
| -------------------------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| `.proto` files                   | **No**             | We pass ts-proto generated `packageDefinition`; Nest never calls `loadSync(protoPath)`                                  |
| `@grpc/proto-loader` **package** | **Yes**            | `ClientGrpcProxy` / `ServerGrpc` `require()` it eagerly in their constructors, used or not; missing → `process.exit(1)` |

## How we configure gRPC

`apps/document/src/{note,authorization}/*.module.ts`:

```ts
transport: Transport.GRPC,
options: {
  package: NOTE_PACKAGE_NAME,
  packageDefinition: { [`${NOTE_PACKAGE_NAME}.${NOTE_SERVICE_NAME}`]: NoteServiceService },
  url: servicesCfg.noteUrl,
},
```

`NoteServiceService` comes from `@notopia-uit/pb` (ts-proto, `outputServices=grpc-js`), a pure-JS
service definition with encode/decode functions. No protobuf schema is parsed at runtime.

## What Nest does (`@nestjs/microservices@11.2.7`)

```js
// client/client-grpc.js — constructor
grpcPackage            = loadPackage('@grpc/grpc-js', ...);
grpcProtoLoaderPackage = loadPackage(protoLoader, ..., () => require('@grpc/proto-loader')); // always
this.grpcClients = this.createClients();   // → loadProto()

// helpers/grpc-helpers.js
function getGrpcPackageDefinition(options, grpcProtoLoaderPackage) {
  ...
  return packageDefinition || grpcProtoLoaderPackage.loadSync(file, options.loader); // loader only if no packageDefinition
}

// @nestjs/common/utils/load-package.util.js
catch (e) { logger.error(`The "${name}" package is missing...`); process.exit(1); }
```

`server/server-grpc.js` has the same eager require. `@grpc/proto-loader` is **not** a peer dep of
`@nestjs/microservices`. It resolves today because `@grpc/grpc-js` depends on it and pnpm's hidden hoist
(`node_modules/.pnpm/node_modules`) makes it reachable. `apps/document` also declares it directly
(since `refactor: move from connectrpc to grpc`), which is what guarantees it lands in the pruned Docker
install. **Keep that direct dep**, even though our code never imports it.

## Upstream history

- [nestjs/nest#8457](https://github.com/nestjs/nest/issues/8457): "provide a packageDefinition option
  instead of providing .proto files". Before this, Nest _did_ require `.proto` files.
  - [#8465](https://github.com/nestjs/nest/pull/8465) (2021-11): `packageDefinition` for `ClientGrpc`.
  - [#10530](https://github.com/nestjs/nest/pull/10530) (2023-12): `packageDefinition` for `ServerGrpc`.
- [#12560](https://github.com/nestjs/nest/issues/12560) (closed 2024-11): proto-loader options are
  server-global rather than per-package (related, not our problem).
- No issue found about making the `@grpc/proto-loader` require lazy when `packageDefinition` is used.

## Verified (2026-10)

- A real `ClientGrpcProxy` with `NoteServiceService`, plus a preloaded hook logging every
  `fs.*('*.proto')`, then a `getNoteName` call: it reached the network (`code=14` UNAVAILABLE) with
  **zero** `.proto` accesses. Positive control: `proto-loader.loadSync('proto/note/note.proto')` under
  the same hook logged the reads.
- Inside the `document` image, `@grpc/proto-loader` and `@grpc/grpc-js` resolve from
  `@nestjs/microservices`.
- `/proto/` is in `.dockerignore`, and that's fine. Only switching a module to `protoPath` would require
  shipping `.proto` files (plus `buf/validate`, `google/protobuf` imports).
