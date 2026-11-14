#!/usr/bin/env bash
set -euo pipefail

projects=(web note authorization billing document search-worker)
inputs=(INPUT_RELEASE_WEB INPUT_RELEASE_NOTE INPUT_RELEASE_AUTHORIZATION INPUT_RELEASE_BILLING INPUT_RELEASE_DOC INPUT_RELEASE_SW)
selected=()

case "${EVENT_NAME:-workflow_dispatch}" in
workflow_dispatch)
  for input in INPUT_RELEASE_ALL "${inputs[@]}"; do
    case "${!input:-false}" in
    true | false | '') ;;
    *)
      echo "Invalid boolean for $input: ${!input}" >&2
      exit 1
      ;;
    esac
  done
  for i in "${!projects[@]}"; do
    input="${inputs[$i]}"
    if [[ "${INPUT_RELEASE_ALL:-false}" == true || "${!input:-false}" == true ]]; then
      selected+=("${projects[$i]}")
    fi
  done
  ;;
push)
  # Dormant until SHOULD_RELEASE explicitly enables pushes to main.
  affected="$(nx show projects --affected --with-target docker:build --json)"
  for project in "${projects[@]}"; do
    if jq -e --arg project "$project" 'index($project) != null' <<<"$affected" >/dev/null; then
      selected+=("$project")
    fi
  done
  ;;
*)
  echo "Unsupported release event: $EVENT_NAME" >&2
  exit 1
  ;;
esac

IFS=,
echo "${selected[*]}"
