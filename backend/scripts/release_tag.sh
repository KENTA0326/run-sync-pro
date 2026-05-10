#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
backend_dir="$repo_root/backend"
version_file="$backend_dir/VERSION"

if [[ ! -f "$version_file" ]]; then
  echo "VERSION file not found: $version_file" >&2
  exit 1
fi

version="$(tr -d '[:space:]' < "$version_file")"
"$backend_dir/scripts/semver_check.sh" "$version" >/dev/null
tag="v$version"

if git -C "$repo_root" rev-parse "$tag" >/dev/null 2>&1; then
  echo "tag already exists: $tag" >&2
  exit 1
fi

git -C "$repo_root" tag -a "$tag" -m "release $tag"
echo "created tag: $tag"
echo "push with: git push origin $tag"
