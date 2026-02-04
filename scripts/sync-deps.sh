#!/usr/bin/env bash

# post-checkout arguments:
# $1 = previous HEAD, $2 = current HEAD, $3 = branch flag (1 for branch, 0 for file)
PREV_HEAD=${1:-"HEAD@{1}"}
NEW_HEAD=${2:-"HEAD"}
IS_BRANCH_CHECKOUT=${3:-1}

# Only run if we actually switched branches/commits
if [ "$IS_BRANCH_CHECKOUT" -eq 0 ]; then
  exit 0
fi

# Get changed files once
CHANGED_FILES=$(git diff --name-only "$PREV_HEAD" "$NEW_HEAD")

# Detection via logic
PNPM_SYNC=false
GO_SYNC=false
MISE_SYNC=false

if echo "$CHANGED_FILES" | grep -q "pnpm-lock.yaml"; then PNPM_SYNC=true; fi
if echo "$CHANGED_FILES" | grep -q "go.sum"; then GO_SYNC=true; fi
if echo "$CHANGED_FILES" | grep -qE ".*mise.*\.toml"; then MISE_SYNC=true; fi

# Exit early if nothing to do
if ! $PNPM_SYNC && ! $GO_SYNC && ! $MISE_SYNC; then
  exit 0
fi

echo "🔔 Dependency changes detected."

# Ensure we have a TTY for the prompt
if [ -t 0 ] || [ -c /dev/tty ]; then
  read -p "Sync dependencies now? [Y/n] " -n 1 -r </dev/tty
  echo
  if [[ ! $REPLY =~ ^[Yy]$ && ! -z $REPLY ]]; then
    echo "Aborted."
    exit 0
  fi
fi

# 1. Sync Mise first (tooling provider)
if [ "$MISE_SYNC" = true ]; then
  echo "🚀 mise install..."
  mise install
fi

# 2. Sync Languages
if [ "$PNPM_SYNC" = true ]; then
  echo "📦 pnpm install..."
  pnpm install
fi

if [ "$GO_SYNC" = true ]; then
  echo "🐹 go mod download..."
  go mod download
fi

echo "✅ Environment synced."
