# cmd/note

Entrypoint only — `main.go`, `metadata.go` (service name `notopia-note`), Wire
definition. Logic in [`internal/note/`](../../internal/note/AGENTS.md).

After editing `wire.go`, run `nx wire note` (never hand-edit `wire_gen.go`).