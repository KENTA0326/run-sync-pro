#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 2
fi

version="$1"
re='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[-0-9A-Za-z.]+)?(\+[0-9A-Za-z.-]+)?$'

if [[ ! "$version" =~ $re ]]; then
  echo "invalid semver: $version" >&2
  exit 1
fi

echo "valid semver: $version"
