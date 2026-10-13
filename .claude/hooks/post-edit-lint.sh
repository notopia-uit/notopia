#!/usr/bin/env bash
set -u

file=$(jq -r '.tool_input.file_path // empty')
[ -n "$file" ] || exit 0

run_if_found() {
  local bin=$1
  shift
  if [ -x "$bin" ] || command -v "$bin" >/dev/null 2>&1; then
    "$bin" "$@"
  fi
}

oxlint=$CLAUDE_PROJECT_DIR/node_modules/.bin/oxlint
oxfmt=$CLAUDE_PROJECT_DIR/node_modules/.bin/oxfmt

case "$file" in
  *.js | *.jsx | *.ts | *.tsx)
    run_if_found "$oxlint" --fix "$file"
    run_if_found "$oxfmt" "$file"
    ;;
  *.go)
    run_if_found gofumpt -w "$file"
    run_if_found golangci-lint run --fix "$file"
    ;;
esac

exit 0
