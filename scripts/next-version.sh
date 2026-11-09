#!/usr/bin/env bash
set -euo pipefail

# Usage: scripts/next-version.sh
# Returns the next patch version for prefix "pakehub".
# Versions before 1.2.0 belong to the old release line. The Bun migration
# starts the new release line at 1.2.0.

PREFIX="pakehub"
MIN_MAJOR=1
MIN_MINOR=2
MIN_PATCH=0

# Fetch tags once
git fetch --tags origin --force >/dev/null 2>&1

LATEST_TAG=$(git tag -l "${PREFIX}-[0-9]*" --sort=v:refname | tail -1 || echo "")

if [ -z "$LATEST_TAG" ]; then
  echo "${MIN_MAJOR}.${MIN_MINOR}.${MIN_PATCH}"
else
  VERSION=${LATEST_TAG#${PREFIX}-}
  IFS='.' read -r major minor patch <<< "$VERSION"
  if [ -z "$patch" ]; then patch=0; fi
  if (( major < MIN_MAJOR || (major == MIN_MAJOR && minor < MIN_MINOR) )); then
    echo "${MIN_MAJOR}.${MIN_MINOR}.${MIN_PATCH}"
  else
    NEW_PATCH=$((patch + 1))
    echo "${major}.${minor}.${NEW_PATCH}"
  fi
fi
