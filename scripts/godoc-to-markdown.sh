#!/usr/bin/env bash

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
GOVALIDATOR_DIR="${SCRIPT_DIR}/../docs/godoc/govalidator.md"
mkdir -p "$(dirname "$GOVALIDATOR_DIR")"

TEMP_HTML=$(mktemp --suffix=.html)

curl -s 'https://pkg.go.dev/github.com/go-playground/validator/v10' | htmlq '.UnitDoc' >"$TEMP_HTML"

markitdown "$TEMP_HTML" >"$GOVALIDATOR_DIR"

rm "$TEMP_HTML"
