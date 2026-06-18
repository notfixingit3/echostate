#!/usr/bin/env bash
# Remove stale EchoState local build images and Docker build cache.
#
# Keeps pulled/running GHCR images (ghcr.io/notfixingit3/echostate*:beta|main).
# Safe to run while the stack is up — only deletes unused local tags.
#
# Usage:
#   ./scripts/docker-clean.sh          # images + build cache
#   ./scripts/docker-clean.sh --dry-run

set -euo pipefail

dry_run=false
if [[ "${1:-}" == "--dry-run" ]]; then
  dry_run=true
fi

run() {
  if $dry_run; then
    printf '[dry-run] '
    printf '%q ' "$@"
    printf '\n'
  else
    "$@"
  fi
}

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

echo "Docker disk usage (before):"
docker system df

echo
echo "Removing unused EchoState local build images…"
patterns=(
  'echostate-api:local'
  'echostate-frontend:local'
  'echostate-api:latest'
  'echostate-frontend:latest'
  'echostate-api:test'
  'echostate-frontend:test'
  'echostate-frontend-test:latest'
)

for pattern in "${patterns[@]}"; do
  ids="$(docker images --format '{{.ID}} {{.Repository}}:{{.Tag}}' | awk -v p="$pattern" '$2 == p {print $1}')"
  if [[ -n "$ids" ]]; then
    while read -r id; do
      [[ -z "$id" ]] && continue
      run docker rmi "$id"
    done <<< "$ids"
  fi
done

echo
echo "Pruning dangling images…"
run docker image prune -f

echo
echo "Pruning build cache (this is usually the big one)…"
run docker builder prune -af

echo
echo "Docker disk usage (after):"
docker system df