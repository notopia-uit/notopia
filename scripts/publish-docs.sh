#!/usr/bin/env bash
set -euo pipefail

remote="$1"
output="$(realpath "$2")"
checkout="$3"

if [[ -e "$output/pr" || -L "$output/pr" ]]; then
  echo "Production output must not contain the reserved pr directory" >&2
  exit 1
fi

git clone --single-branch --branch gh-pages "$remote" "$checkout"
find "$checkout" -mindepth 1 -maxdepth 1 ! -name .git ! -name pr -exec rm -rf -- {} +
cp -a "$output/." "$checkout/"
git -C "$checkout" add -A
if git -C "$checkout" diff --cached --quiet; then
  echo "Production documentation is unchanged"
  exit 0
fi
git -C "$checkout" commit -m "docs: deploy production"
git -C "$checkout" push origin gh-pages
