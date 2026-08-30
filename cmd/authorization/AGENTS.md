# Authorization Service Entrypoint

This directory contains the service entrypoint for the Authorization service.

## Files

| File | Purpose |
|------|---------|
| `main.go` | Service entrypoint — sets up logging, signal handling, initializes server via Wire, and runs the server |
| `wire.go` | Wire dependency injection definition — combines `authorization.ProviderSet` with service metadata to build `InitializeServer` |
| `wire_gen.go` | Auto-generated Wire code — do not edit manually |
| `metadata.go` | Service name (`notopia-authorization`) and version constants |

## Dependency Injection

Wire is used for compile-time dependency injection. The `InitializeServer` function wires together all internal packages from `internal/authorization/`.

To regenerate `wire_gen.go` after modifying `wire.go`:

```sh
nx wire authorization
```

## Source Code

All business logic, policy engine, controllers, and infrastructure implementations live in `internal/authorization/`. See [`../../internal/authorization/AGENTS.md`](../../internal/authorization/AGENTS.md) for details.
