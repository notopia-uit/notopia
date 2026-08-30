# cmd/note

Entrypoint only — `main.go`, `metadata.go` (service name `notopia-note`), and the Wire
definition. All logic is in [`internal/note/`](../../internal/note/AGENTS.md).

After editing `wire.go`, regenerate with `nx wire note` (never hand-edit `wire_gen.go`).
