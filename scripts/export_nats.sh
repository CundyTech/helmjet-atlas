#!/usr/bin/env bash
set -euo pipefail

# export_nats.sh
# Exports NATS JetStream streams and consumers into a single JSON file suitable for upload
# Usage: ./scripts/export_nats.sh [CLUSTER_NAME] [OUTFILE]
# Example: ./scripts/export_nats.sh default visualization/examples/nats_upload.json

CLUSTER_NAME="${1:-default}"
OUTFILE="${2:-nats_upload.json}"
TMP_STREAMS="$(mktemp)"
TMP_OUT="$(mktemp)"
trap 'rm -f "$TMP_STREAMS" "$TMP_OUT"' EXIT

if ! command -v nats >/dev/null 2>&1; then
  echo "nats CLI not found in PATH. Install https://docs.nats.io/tools/natscli" >&2
  exit 2
fi

echo "Gathering stream list..."
nats stream ls --json > "$TMP_STREAMS"

echo -n "{" > "$OUTFILE"
echo -n '"streams":[' >> "$OUTFILE"
first=true
jq -r '.[].name' "$TMP_STREAMS" | while read -r stream; do
  info=$(nats stream info "$stream" --json 2>/dev/null || echo "null")
  if [ "$info" = "null" ]; then
    continue
  fi
  # inject cluster field
  info_with_cluster=$(echo "$info" | jq --arg c "$CLUSTER_NAME" '. + {cluster: $c}')
  if [ "$first" = "true" ]; then
    echo -n "$info_with_cluster" >> "$OUTFILE"
    first=false
  else
    echo -n "," >> "$OUTFILE"
    echo -n "$info_with_cluster" >> "$OUTFILE"
  fi
done
echo -n '],"consumers":[' >> "$OUTFILE"
first=true
jq -r '.[].name' "$TMP_STREAMS" | while read -r stream; do
  cmap=$(nats consumer ls "$stream" --json 2>/dev/null || echo "[]")
  names=$(echo "$cmap" | jq -r '.[].name')
  for cn in $names; do
    info=$(nats consumer info --json "$stream" "$cn" 2>/dev/null || echo "null")
    if [ "$info" = "null" ]; then
      continue
    fi
    info_with_cluster=$(echo "$info" | jq --arg c "$CLUSTER_NAME" '. + {cluster: $c}')
    if [ "$first" = "true" ]; then
      echo -n "$info_with_cluster" >> "$OUTFILE"
      first=false
    else
      echo -n "," >> "$OUTFILE"
      echo -n "$info_with_cluster" >> "$OUTFILE"
    fi
  done
done
echo -n ']}' >> "$OUTFILE"

echo "Wrote $OUTFILE"
