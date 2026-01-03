# Kubernetes Integration

This document describes the HTTP API and model mappings used by the Kubernetes integration in Helmjet Atlas.

## Overview

The Kubernetes integration connects to a Kubernetes cluster (in-cluster or via kubeconfig file) and syncs Service objects into the application's `Microservice` model.

Key interactions with Kubernetes:

- List services: `clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})`

The server exposes a simple HTTP endpoint to trigger a one-shot sync.

## API Endpoint

- POST `/api/v1/integrations/k8s/sync`

Request body (JSON, optional):

```
{
  "kubeconfig": "/path/to/kubeconfig",   // optional, if running out-of-cluster
  "namespaces": ["default", "kube-system"] // optional list to restrict the sync
}
```

Alternatively the request may set the `X-KUBECONFIG` header to pass a kubeconfig path.

Response (200):

```
{
  "message": "k8s sync triggered"
}
```

On error, a 500 is returned with the error in the body.

## Model mappings

Data discovered from Kubernetes Services is mapped into the `Microservice` model (see `internal/models/models.go`). The mapping used in the sync is:

- `name` (string): Kubernetes Service metadata.name
- `namespace` (string): Kubernetes Service metadata.namespace
- `cluster` (string): left empty by default; can be set from config
- `ports` ([]Port): maps Service ports into `Port` objects with fields `name`, `containerPort` (the service port), and `protocol`.
- `updatedAt`, `lastSynced`: timestamps set during the sync.

Example JSON produced when stored:

```json
{
  "name": "orders-api",
  "namespace": "default",
  "cluster": "",
  "ports": [ { "name": "http", "containerPort": 8080, "protocol": "TCP" } ],
  "lastSyncedAt": "2025-12-20T12:34:56Z"
}
```

## Namespaces filtering

If the `namespaces` array is provided in the request body, only Services whose `metadata.namespace` matches an entry in the list will be upserted. If the list is empty or omitted, all namespaces are scanned.

## Notes

- The integration attempts in-cluster configuration first; if unavailable it will fall back to the provided kubeconfig path.
- The sync is one-shot and uses the repository's `UpsertByNameNamespace` to update or insert Services.
- You should ensure the service account or kubeconfig used has permission to `list` Services across the target namespaces.
- You should ensure the service account or kubeconfig used has permission to `list` Services across the target namespaces.

## Vendor documentation

- Kubernetes API reference: https://kubernetes.io/docs/reference/kubernetes-api/
- Services API (example): https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#service-v1-core
- `client-go` package docs: https://pkg.go.dev/k8s.io/client-go
- Kubeconfig / organizing access docs: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/

Refer to these vendor sources for details on Service fields, RBAC requirements, and kubeconfig usage.
