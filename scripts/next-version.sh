#!/usr/bin/env bash
set -euo pipefail

# Usage: scripts/next-version.sh <slug>
# Returns the next patch version based on existing tags.

SLUG="$1"
if [ -z "$SLUG" ]; then
  echo "❌ Missing slug argument" >&2
  exit 1
fi

# Fetch tags once
git fetch --tags origin --force >/dev/null 2>&1

# Use git tag with version sorting (more efficient than sort -V)
LATEST_TAG=$(git tag --list "${SLUG}-[0-9]*" --sort=v:refname | tail -1 || echo "")

if [ -z "$LATEST_TAG" ]; then
  echo "1.1.0"
else
  VERSION=${LATEST_TAG#${SLUG}-}
  IFS='.' read -r major minor patch <<< "$VERSION"
  if [ -z "$patch" ]; then patch=0; fi
  NEW_PATCH=$((patch + 1))
  echo "${major}.${minor}.${NEW_PATCH}"
fi