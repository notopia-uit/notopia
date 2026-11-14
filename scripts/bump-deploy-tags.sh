#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 ]]; then
  echo 'Usage: bump-deploy-tags.sh <head-sha> <comma-list>' >&2
  exit 1
fi
head="$(git rev-parse --verify --end-of-options "$1^{commit}")"
if [[ "$(git rev-parse HEAD)" != "$head" ]]; then
  echo 'Release HEAD changed; refusing to update deployment values' >&2
  exit 1
fi
[[ -n "$2" ]] || exit 0
: "${RELEASE_TAGS_BEFORE:?Set RELEASE_TAGS_BEFORE to the tag snapshot taken before this release}"
[[ -f "$RELEASE_TAGS_BEFORE" ]] || {
  echo 'Missing pre-release tag snapshot' >&2
  exit 1
}
if [[ "$2" == ,* || "$2" == *, || "$2" == *,,* ]]; then
  echo 'Invalid release project list' >&2
  exit 1
fi

IFS=, read -r -a projects <<<"$2"
declare -A seen=()
files=()
versions=()
new_tags="$(LC_ALL=C comm -13 <(LC_ALL=C sort "$RELEASE_TAGS_BEFORE") <(git tag --list | LC_ALL=C sort))"
head_tags="$(git tag --points-at "$head")"
for project in "${projects[@]}"; do
  case "$project" in
  web | note | authorization | billing | document | search-worker) ;;
  *)
    echo "Unknown release project: $project" >&2
    exit 1
    ;;
  esac
  if [[ -v "seen[$project]" ]]; then
    echo "Duplicate release project: $project" >&2
    exit 1
  fi
  seen[$project]=1
  tags=()
  while IFS= read -r tag; do
    if [[ "$tag" == "$project@"* ]] && grep -Fxq -- "$tag" <<<"$head_tags"; then
      tags+=("$tag")
    fi
  done <<<"$new_tags"
  if [[ "${#tags[@]}" != 1 ]]; then
    echo "Expected exactly one new $project@<version> tag at $head; found ${#tags[@]}" >&2
    exit 1
  fi
  version="${tags[0]#*@}"
  if [[ ! "$version" =~ ^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,127}$ ]]; then
    echo "Invalid Docker tag for $project: $version" >&2
    exit 1
  fi
  file="deploy/k8s/clusters/homelab/apps/$project/values.yaml"
  if [[ ! -f "$file" ]]; then
    echo "::warning::No deployment values for $project at $file; skipping" >&2
    continue
  fi
  git ls-files --error-unmatch -- "$file" >/dev/null
  git diff --exit-code HEAD -- "$file" >/dev/null
  yq -e '.image | kind == "map"' "$file" >/dev/null
  files+=("$file")
  versions+=("$version")
done

for i in "${!files[@]}"; do
  VERSION="${versions[$i]}" yq -i '.image.tag = strenv(VERSION)' "${files[$i]}"
  echo "Updated ${files[$i]} to ${versions[$i]}"
done
