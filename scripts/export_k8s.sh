#!/usr/bin/env bash
set -euo pipefail

# export_k8s.sh
# Exports Kubernetes Services and Deployments into a microservice upload JSON
# Usage: ./scripts/export_k8s.sh [CLUSTER_NAME] [OUTFILE]
# Example: ./scripts/export_k8s.sh minikube visualization/examples/microservices_upload.json

CLUSTER_NAME="${1:-local}"
OUTFILE="${2:-microservices_upload.json}"

TMP_SV="$(mktemp)"
TMP_DP="$(mktemp)"
trap 'rm -f "$TMP_SV" "$TMP_DP"' EXIT

echo "Fetching Services and Deployments..."
kubectl get services --all-namespaces -o json > "$TMP_SV"
kubectl get deployments --all-namespaces -o json > "$TMP_DP"

# Build upload array. We try to enrich service with a deployment of the same name+namespace (common pattern).
# Fields produced match internal/models.Microservice expected by the upload endpoint.

jq -n --arg cluster "$CLUSTER_NAME" --slurpfile sv "$TMP_SV" --slurpfile dp "$TMP_DP" '
  ($sv[0].items) as $services |
  ($dp[0].items) as $deploys |
  [ $services[] | . as $s |
      (
        ($deploys[]? | select(.metadata.namespace == $s.metadata.namespace and .metadata.name == $s.metadata.name))? as $d
      ) |
      {
        name: $s.metadata.name,
        namespace: $s.metadata.namespace,
        cluster: $cluster,
        image: ($d.spec.template.spec.containers[0].image // null),
        replicas: ($d.spec.replicas // null),
        status: ($d.status?.conditions? // null),
        labels: ($s.metadata.labels // {}),
        ports: ($s.spec.ports // [] | map({name: (.name // ""), containerPort: (.port), protocol: (.protocol // "TCP")}))
      }
  ]
' > "$OUTFILE"

echo "Wrote $OUTFILE"
