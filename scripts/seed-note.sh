#!/usr/bin/env bash

set -euo pipefail

# PostgreSQL connection parameters with fallback to env vars or defaults
PG_HOST="${PG_HOST:-${PGHOST:-localhost}}"
PG_PORT="${PG_PORT:-${PGPORT:-5432}}"
PG_USER="${PG_USER:-${PGUSER:-postgres}}"
PG_PASSWORD="${PG_PASSWORD:-${PGPASSWORD:-}}"
PG_DATABASE="${PG_DATABASE:-${PGDATABASE:-notopia}}"

# Parse command-line arguments
# Usage: ./scripts/seed-note.sh [--host HOST] [--port PORT] [--user USER] [--password PASSWORD] [--db DATABASE]
while [[ $# -gt 0 ]]; do
  case "$1" in
    --host)
      PG_HOST="$2"
      shift 2
      ;;
    --port)
      PG_PORT="$2"
      shift 2
      ;;
    --user)
      PG_USER="$2"
      shift 2
      ;;
    --password)
      PG_PASSWORD="$2"
      shift 2
      ;;
    --db|--database)
      PG_DATABASE="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
done

# Validate required parameters
if [[ -z "$PG_HOST" ]] || [[ -z "$PG_USER" ]] || [[ -z "$PG_DATABASE" ]]; then
  echo "Error: Missing required PostgreSQL connection parameters" >&2
  echo "Usage: $0 [--host HOST] [--port PORT] [--user USER] [--password PASSWORD] [--db DATABASE]" >&2
  exit 1
fi

# Seed file path (relative to script location)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
SEED_FILE="$REPO_ROOT/internal/notecreateseed/seed.sql"

if [[ ! -f "$SEED_FILE" ]]; then
  echo "Error: Seed file not found: $SEED_FILE" >&2
  echo "Run 'go run ./cmd/notecreateseed' to generate it first" >&2
  exit 1
fi

# Determine SQL client: prefer pgcli, fallback to psql
SQL_CLIENT=""
if command -v pgcli &> /dev/null; then
  SQL_CLIENT="pgcli"
elif command -v psql &> /dev/null; then
  SQL_CLIENT="psql"
else
  echo "Error: Neither pgcli nor psql found. Install one of them to continue." >&2
  exit 1
fi

echo "🌱 Seeding note service..."
echo "  Client: $SQL_CLIENT"
echo "  Host: $PG_HOST"
echo "  Port: $PG_PORT"
echo "  User: $PG_USER"
echo "  Database: $PG_DATABASE"
echo ""

# Build connection string and execute seed
export PGHOST="$PG_HOST"
export PGPORT="$PG_PORT"
export PGUSER="$PG_USER"
export PGDATABASE="$PG_DATABASE"

if [[ -n "$PG_PASSWORD" ]]; then
  export PGPASSWORD="$PG_PASSWORD"
fi

if [[ "$SQL_CLIENT" == "pgcli" ]]; then
  pgcli --file "$SEED_FILE"
else
  psql -f "$SEED_FILE"
fi

echo "✅ Note service seeded successfully!"
