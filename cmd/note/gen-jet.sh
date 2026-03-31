#!/usr/bin/env bash

set -e

CONTAINER_NAME='notopia-note-jet-gen-db'
DB_NAME='pgjet'
DB_USER='postgres'
DB_PASS='postgres'
DB_PORT='15433'
OUTPUT_DIR="../../internal/note/infra/persistence/"

echo "🚀 Starting ephemeral Postgres $CONTAINER_NAME on port $DB_PORT..."
CONTAINER_ID=$(docker run --rm -d \
  --name "$CONTAINER_NAME" \
  -e POSTGRES_PASSWORD=$DB_PASS \
  -e POSTGRES_DB=$DB_NAME \
  -p $DB_PORT:5432 \
  postgres:18.1-alpine3.23)

cleanup() {
  echo "🧹 Cleaning up container..."
  docker stop "$CONTAINER_ID" >/dev/null
}
trap cleanup EXIT

echo "Wait for Postgres to be ready..."
until docker exec "$CONTAINER_ID" pg_isready -U "$DB_USER" >/dev/null 2>&1; do
  sleep 1
done

echo "📦 Applying migrations..."
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=127.0.0.1 port=$DB_PORT user=$DB_USER password=$DB_PASS dbname=$DB_NAME sslmode=disable"
export GOOSE_MIGRATION_DIR=../../internal/note/infra/persistence/pgmigration/
goose up

echo "✈️  Generating Jet code..."
jet -source=postgres \
  -host=127.0.0.1 \
  -port="$DB_PORT" \
  -user="$DB_USER" \
  -password="$DB_PASS" \
  -dbname="$DB_NAME" \
  -schema=public \
  -path="$OUTPUT_DIR"

echo "✅ Success! Code generated in $OUTPUT_DIR"
