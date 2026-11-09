# cmd/authorization

Entrypoint only: `main.go`, `metadata.go` (service name `notopia-authorization`), Wire definition. Logic in
[`internal/authorization/`](../../internal/authorization/AGENTS.md).

After editing `wire.go`, regenerate with `nx wire authorization`; never hand-edit
`wire_gen.go`.