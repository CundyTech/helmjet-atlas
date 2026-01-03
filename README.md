# HelmJet Atlas - Monorepo

Complete solution for visualizing Kubernetes microservices topology and NATS JetStream message bus architecture.

## Overview

HelmJet Atlas is organized as a Go monorepo with three independent modules that work together:

1. **API** (`api/`) - REST API for managing topology data
2. **NATS Integration** (`nats-integration/`) - Watches and syncs NATS JetStream
3. **Kubernetes Integration** (`k8s-integration/`) - Watches and syncs Kubernetes services

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│          Kubernetes Clusters & NATS Clusters            │
└────┬────────────────────────────────────┬────────────────┘
     │                                    │
     ↓                                    ↓
┌──────────────────────┐      ┌────────────────────────┐
│  k8s-integration     │      │  nats-integration      │
│  (watches K8s)       │      │  (watches NATS)        │
└──────────────────────┘      └────────────────────────┘
     │                                    │
     └────────────────┬───────────────────┘
                      ↓
              ┌──────────────────┐
              │     MongoDB      │
              │   (data store)   │
              └──────────────────┘
                      ↑
              ┌──────────────────┐
              │   API Module     │
              │  (REST endpoints)│
              └──────────────────┘
```

## Directory Structure

```
helmjet-atlas/
├── api/                      # REST API module
│   ├── main.go               # API server
│   ├── models.go             # Shared data structures
│   ├── go.mod                # API module definition
│   └── README.md             # API documentation
│
├── nats-integration/         # NATS watcher module
│   ├── main.go               # NATS watcher
│   ├── go.mod                # NATS module definition
│   └── README.md             # NATS documentation
│
├── k8s-integration/          # Kubernetes watcher module
│   ├── main.go               # K8s watcher
│   ├── go.mod                # K8s module definition
│   └── README.md             # K8s documentation
│
├── visualization/            # Web topology dashboard
│   ├── index.html            # Main HTML dashboard
│   ├── topology.js           # D3.js visualization
│   └── README.md             # Visualization documentation
│
├── docs/                     # Documentation
│   ├── GETTING_STARTED.md
│   ├── DEPLOYMENT.md
│   ├── API_ENDPOINTS.md
│   └── [more docs]
│
├── Dockerfile                # API Docker image
├── docker-compose.yml        # Local development stack
├── openapi.yaml             # OpenAPI specification
├── examples.rest            # API examples
├── go.work                  # Monorepo workspace
└── README.md                # This file
```

## Getting Started

### Prerequisites

- Go 1.21+
- MongoDB 4.4+
- Docker & Docker Compose (for containerized setup)
- kubectl (for Kubernetes integration)

### Quick Start - All Services with Docker Compose

```bash
# Clone and navigate to repo
cd helmjet-atlas

# Start all services (API, MongoDB, NATS, K8s watcher)
docker-compose up

# API available at http://localhost:8080
# MongoDB available at localhost:27017
```

### Development Setup

#### Start MongoDB Locally

```bash
docker run -d -p 27017:27017 --name helmjet-mongodb mongo:latest
```

#### Run API Module

```bash
cd api

# Set environment variables
export MONGO_URI=mongodb://localhost:27017
export MONGO_DB=helmjet
export PORT=8080

# Run API server
go run main.go
```

Navigate to `http://localhost:8080/health` to verify the API is running.

#### Run NATS Integration (optional)

```bash
cd nats-integration

# Set environment variables
export NATS_URL=nats://your-nats-cluster:4222
export MONGO_URI=mongodb://localhost:27017
export MONGO_DB=helmjet

# Run NATS watcher
go run main.go
```

#### Run Kubernetes Integration (optional)

```bash
cd k8s-integration

# For in-cluster operation (running in Kubernetes pod):
# No kubeconfig needed, uses service account

# For development with kubeconfig:
export KUBECONFIG=$HOME/.kube/config
export MONGO_URI=mongodb://localhost:27017
export MONGO_DB=helmjet

# Run K8s watcher
go run main.go
```

## Module Details

### API Module (`api/`)

REST API for managing topology data. 26 endpoints across 5 resource types.

**Quick Facts:**
- Framework: Gin
- Port: Configurable via `PORT` env var (default 8080)
- Database: MongoDB
- Endpoints: 26 total (5 resources × 5 operations + 1 health)

**Key Features:**
- Health check endpoint
- CRUD operations for all topology entities
- RESTful API design
- OpenAPI specification

**Resources:**
- Microservices
- NATS Streams
- NATS Consumers
- Service-Stream Links
- Consumer-Service Links

See [api/README.md](./api/README.md) for detailed endpoint documentation.

### NATS Integration Module (`nats-integration/`)

Watches NATS JetStream clusters and syncs streams & consumers to MongoDB.

**Quick Facts:**
- Polling interval: 30 seconds
- Scope: All NATS JetStream streams and consumers
- Database: MongoDB (shared with API)

**Features:**
- Automatic stream discovery
- Consumer tracking
- Metrics collection
- Graceful shutdown

See [nats-integration/README.md](./nats-integration/README.md) for implementation details.

### Kubernetes Integration Module (`k8s-integration/`)

Watches Kubernetes clusters and syncs services to MongoDB.

**Quick Facts:**
- Polling interval: 30 seconds
- Scope: All Kubernetes services across namespaces
- Database: MongoDB (shared with API)

**Features:**
- In-cluster and external cluster support
- Service discovery
- Pod readiness tracking
- Graceful shutdown

**RBAC Requirements:**
```yaml
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list", "watch"]
```

See [k8s-integration/README.md](./k8s-integration/README.md) for implementation details.

## Visualization Dashboard

A modern, interactive web-based topology visualization dashboard for monitoring your system.

**Features:**
- Interactive D3.js force-directed graph
- Real-time visualization of services, streams, and consumers
- Click nodes to view detailed information
- Zoom, pan, and reset layout controls
- Auto-refresh every 30 seconds
- Dark theme optimized for monitoring

**Quick Start:**

```bash
# Option 1: Open directly in browser
open visualization/index.html

# Option 2: Use local web server (recommended)
cd visualization
python -m http.server 8000
# Then visit http://localhost:8000
```

Configure the API endpoint in the sidebar (default: localhost:8080).

See [visualization/README.md](./visualization/README.md) for full documentation.

## Data Model

All modules share a common MongoDB data store with the following collections:

- `microservices` - Kubernetes services
- `nats_streams` - NATS JetStream streams
- `nats_consumers` - NATS JetStream consumers
- `service_stream_links` - Relationships between services and streams
- `consumer_service_links` - Relationships between consumers and services
- `metrics` - Performance metrics
- `clusters` - Cluster registry

See [docs/MONGODB_SCHEMA.md](./docs/MONGODB_SCHEMA.md) for detailed schema.

## API Quick Reference

### Health Check

```
GET /health
```

### Microservices

```
GET    /api/v1/microservices              # List all
POST   /api/v1/microservices              # Create
GET    /api/v1/microservices/:id          # Get one
PUT    /api/v1/microservices/:id          # Update
DELETE /api/v1/microservices/:id          # Delete
```

### NATS Streams

```
GET    /api/v1/streams                    # List all
POST   /api/v1/streams                    # Create
GET    /api/v1/streams/:id                # Get one
PUT    /api/v1/streams/:id                # Update
DELETE /api/v1/streams/:id                # Delete
```

### NATS Consumers

```
GET    /api/v1/consumers                  # List all
POST   /api/v1/consumers                  # Create
GET    /api/v1/consumers/:id              # Get one
PUT    /api/v1/consumers/:id              # Update
DELETE /api/v1/consumers/:id              # Delete
```

### Service-Stream Links

```
GET    /api/v1/links/service-stream       # List all
POST   /api/v1/links/service-stream       # Create
GET    /api/v1/links/service-stream/:id   # Get one
PUT    /api/v1/links/service-stream/:id   # Update
DELETE /api/v1/links/service-stream/:id   # Delete
```

### Consumer-Service Links

```
GET    /api/v1/links/consumer-service     # List all
POST   /api/v1/links/consumer-service     # Create
GET    /api/v1/links/consumer-service/:id # Get one
PUT    /api/v1/links/consumer-service/:id # Update
DELETE /api/v1/links/consumer-service/:id # Delete
```

For detailed documentation, see [docs/API_ENDPOINTS.md](./docs/API_ENDPOINTS.md).

## Development Workflow

### Using Go Workspace

All modules are coordinated through `go.work`:

```bash
# Build all modules
go build ./api ./nats-integration ./k8s-integration

# Run tests across all modules
go test ./...

# Update dependencies (affects all modules)
go get -u ./...
```

### Making Changes

**To modify the data model:**
1. Edit `api/models.go`
2. Run integration tests to verify
3. Both watchers will use updated model

**To add API endpoints:**
1. Add handler in `api/handlers/`
2. Register route in `api/main.go`
3. Update `openapi.yaml`

**To add NATS sync logic:**
1. Implement in `nats-integration/main.go`
2. Test against live NATS cluster
3. Verify data in MongoDB

**To add Kubernetes sync logic:**
1. Implement in `k8s-integration/main.go`
2. Test with local K8s cluster (minikube)
3. Verify data in MongoDB

## Deployment

### Docker

Build API image:

```bash
docker build -t helmjet-atlas-api:latest .
```

### Kubernetes

```bash
# Create ConfigMap for settings
kubectl create configmap helmjet-config --from-literal=MONGO_DB=helmjet

# Create Secret for credentials
kubectl create secret generic helmjet-mongo --from-literal=MONGO_URI=mongodb://user:pass@host:27017

# Deploy
kubectl apply -f k8s-manifests/
```

### Docker Compose (Local Development)

```bash
docker-compose up
```

Includes:
- API service (port 8080)
- MongoDB (port 27017)
- NATS (port 4222, optional)
- NATS Integration (optional)
- K8s Integration (optional)

## Documentation

- [GETTING_STARTED.md](./docs/GETTING_STARTED.md) - Detailed setup guide
- [API_ENDPOINTS.md](./docs/API_ENDPOINTS.md) - Complete API reference
- [MONGODB_SCHEMA.md](./docs/MONGODB_SCHEMA.md) - Database schema design
- [DEPLOYMENT.md](./docs/DEPLOYMENT.md) - Production deployment guide
- [JETSTREAM_DESIGN.md](./docs/JETSTREAM_DESIGN.md) - NATS architecture details
- [KEY_OPERATIONS_EXPLAINED.md](./docs/KEY_OPERATIONS_EXPLAINED.md) - Core functionality

## Monitoring & Logging

Each module outputs structured logs to stdout:

```
[API] 2024-01-15 10:30:45 INFO Starting server on :8080
[NATS] 2024-01-15 10:30:46 INFO Connected to NATS
[K8s] 2024-01-15 10:30:47 INFO Connected to Kubernetes
```

The API exposes a health endpoint:

```bash
curl http://localhost:8080/health
```

## Contributing

1. Create a feature branch
2. Make changes (following module structure)
3. Test locally with all modules
4. Submit pull request

## Troubleshooting

**MongoDB connection failed:**
- Verify MongoDB is running and accessible
- Check MONGO_URI is correct
- Ensure MONGO_DB environment variable is set

**NATS connection failed:**
- Verify NATS cluster is accessible
- Check NATS_URL is correct
- Ensure credentials if authentication is enabled

**Kubernetes connection failed (k8s-integration):**
- For in-cluster: Verify pod has service account with proper RBAC
- For external: Verify KUBECONFIG points to valid kubeconfig file
- Check cluster endpoint is accessible

**API returns 502 Bad Gateway:**
- Verify all upstream services are running
- Check MongoDB connection
- Review application logs

## License

[Specify your license]

## Support

For issues, questions, or contributions, please open an issue on GitHub.
