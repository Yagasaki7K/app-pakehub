#!/usr/bin/env bash
set -euo pipefail

# Usage: scripts/next-version.sh
# Returns the next patch version for prefix "pakehub".

PREFIX="pakehub"

# Fetch tags once
git fetch --tags origin --force >/dev/null 2>&1

LATEST_TAG=$(git tag -l "${PREFIX}-[0-9]*" --sort=v:refname | tail -1 || echo "")

if [ -z "$LATEST_TAG" ]; then
  echo "1.1.0"
else
  VERSION=${LATEST_TAG#${PREFIX}-}
  IFS='.' read -r major minor patch <<< "$VERSION"
  if [ -z "$patch" ]; then patch=0; fi
  NEW_PATCH=$((patch + 1))
  echo "${major}.${minor}.${NEW_PATCH}"
fi