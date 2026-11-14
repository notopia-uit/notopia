#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 ]]; then
  echo 'Usage: push-deploy-tags.sh <head-sha> <comma-list>' >&2
  exit 1
fi
if [[ "${GITHUB_EVENT_NAME:-}" != workflow_dispatch || "${GITHUB_REF:-}" != refs/heads/main ]]; then
  echo 'Deployment pushes are only allowed for manual releases on main' >&2
  exit 1
fi
head="$(git rev-parse --verify --end-of-options "$1^{commit}")"
if [[ "$(git rev-parse HEAD)" != "$head" ]]; then
  echo 'Release HEAD changed; refusing to push deployment tags' >&2
  exit 1
fi
[[ -n "$2" ]] || exit 0
remote_head="$(git ls-remote --exit-code origin refs/heads/main | cut -f1)"
if [[ "$remote_head" != "$head" ]]; then
  echo 'origin/main advanced during release; refusing deployment commit (rerun from latest main)' >&2
  exit 1
fi
git diff --cached --exit-code >/dev/null
if [[ "$2" == ,* || "$2" == *, || "$2" == *,,* ]]; then
  echo 'Invalid release project list' >&2
  exit 1
fi
IFS=, read -r -a projects <<<"$2"
files=()
releases=()
for project in "${projects[@]}"; do
  case "$project" in
  web | note | authorization | billing | document | search-worker) ;;
  *)
    echo "Unknown release project: $project" >&2
    exit 1
    ;;
  esac
  file="deploy/k8s/clusters/homelab/apps/$project/values.yaml"
  [[ -f "$file" ]] || continue
  if ! git diff --quiet -- "$file"; then
    files+=("$file")
    releases+=("$project@$(yq -r '.image.tag' "$file")")
  fi
done
if [[ "${#files[@]}" == 0 ]]; then
  echo 'No deployment tag changes to commit'
  exit 0
fi
git add -- "${files[@]}"
summary="$(
  IFS=,
  echo "${releases[*]}"
)"
git commit -m "chore(deploy): release $summary"
git -c push.followTags=false push origin HEAD:refs/heads/main
