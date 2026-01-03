Kubernetes manifests for helmjet-atlas monolith

Overview
- Namespace: `helmjet-atlas`
- Monolith deployment uses image `docker.io/dancundy/helmjet-atlas:monolith` (update as needed)
- Includes simple MongoDB and NATS deployments for local/dev clusters

Build, tag and push the image (Docker Hub `dancundy`):

```bash
# from repository root
docker build -t dancundy/helmjet-atlas:monolith .
docker push dancundy/helmjet-atlas:monolith
```

Deploy to cluster (requires `kubectl` and an available cluster):

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -n helmjet-atlas -f k8s/mongo-deployment.yaml
kubectl apply -n helmjet-atlas -f k8s/nats-deployment.yaml
kubectl apply -n helmjet-atlas -f k8s/monolith-deployment.yaml
```

Notes
- The MongoDB deployment here is intentionally minimal (no persistent volume). For production, replace with a StatefulSet and PVC or use a managed DB.
- The Ingress assumes an `nginx` ingress controller and DNS pointing at the controller.
- If you're using a local cluster like `kind`, load the image into the cluster or push to Docker Hub.
