# Note Service Entrypoint

This directory contains the service entrypoint for the Note service.

## Files

| File | Purpose |
|------|---------|
| `main.go` | Service entrypoint — sets up logging, signal handling, initializes server via Wire, and runs the server |
| `wire.go` | Wire dependency injection definition — combines `note.ProviderSet` with service metadata to build `InitializeServer` |
| `wire_gen.go` | Auto-generated Wire code — do not edit manually |
| `metadata.go` | Service name (`notopia-note`) and version constants |

## Dependency Injection

Wire is used for compile-time dependency injection. The `InitializeServer` function wires together all internal packages from `internal/note/`.

To regenerate `wire_gen.go` after modifying `wire.go`:

```sh
nx wire note
```

## Source Code

All business logic, domain models, controllers, and infrastructure implementations live in `internal/note/`. See [`../../internal/note/AGENTS.md`](../../internal/note/AGENTS.md) for details.
