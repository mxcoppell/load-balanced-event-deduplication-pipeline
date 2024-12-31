# Load-Balanced Event Deduplication Pipeline

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This project implements a distributed system to test and analyze key expiration events in Redis, focusing on efficient event processing and deduplication across multiple consumers.

![Web UI](asset/test-ui.jpg)

## System Components

- **Generator Service**: Creates Redis keys with specified TTLs
- **Consumer Service**: Processes key expiration events with deduplication
- **Redis**: Stores keys and provides key expiration notifications
- **NATS**: Handles message distribution using WorkQueue policy for even load distribution

## Architecture

1. The Generator creates keys in Redis with specified TTLs
2. When keys expire, Redis publishes expiration events
3. Consumers receive these events and use a deduplication mechanism to ensure each event is processed only once
4. NATS JetStream with WorkQueue policy ensures even distribution of processed events among consumers

## Prerequisites

- Docker Desktop with Kubernetes enabled
- kubectl
- Go 1.21+ (optional, only needed for local development)
- Node.js 20+ (optional, only needed for local UI development)

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/mxcoppell/load-balanced-event-deduplication-pipeline.git
cd load-balanced-event-deduplication-pipeline
```

### 2. Build the Services

There are two ways to build the services:

#### Option 1: Direct Docker Build (Recommended)

```bash
# Build the services directly using Docker
docker build -t generator:latest -f docker/Dockerfile.generator .
docker build -t consumer:latest -f docker/Dockerfile.consumer .
```

#### Option 2: Local Development Build

Only needed if you want to develop the UI locally:
```bash
cd web && npm install && npm run build && cd ..
docker build -t generator:latest -f docker/Dockerfile.generator .
docker build -t consumer:latest -f docker/Dockerfile.consumer .
```

### 3. Deploy to Kubernetes

Deploy the services in the following order:

```bash
# 1. Deploy Redis and wait for it to be ready
kubectl apply -f k8s/redis/deployment.yaml
kubectl wait --for=condition=ready pod -l app=redis --timeout=60s

# 2. Deploy NATS and wait for it to be ready
kubectl apply -f k8s/nats/deployment.yaml
kubectl wait --for=condition=ready pod -l app=nats --timeout=60s

# 3. Deploy generator and consumer services
kubectl apply -f k8s/generator/deployment.yaml
kubectl apply -f k8s/consumer/deployment.yaml
```

## Troubleshooting

1. If pods are in CrashLoopBackOff state:
   - Check if Redis and NATS are fully ready
   - Use `kubectl logs deployment/<service-name>` to view service logs
   - If needed, restart services: `kubectl rollout restart deployment/<service-name>`

2. If services can't connect:
   - Ensure services are deployed in the correct order (Redis → NATS → Generator → Consumer)
   - Verify all services are running: `kubectl get pods`

3. Common Issues:
   - Generator failing to start: Usually means Redis is not ready
   - Consumer pods restarting: Normal during initial NATS stream creation
   - Connection refused errors: Indicates dependency services are not ready

## Usage

1. Access the WebUI at `http://localhost:30080`

2. Configure test parameters:
   - Number of Keys: Total keys to generate
   - Key Delay (ms): Delay between key generations
   - Key TTL (ms): Time-to-live for each key
   - Dedup Window (ms): Duration for deduplication window

3. Click "Start Test" to begin the test

4. Monitor the results:
   - Generated Keys: Total keys created
   - Consumed Keys: Total keys processed
   - Consumer Metrics: Distribution of processed events across consumers

## Development

### Project Structure

```
.
├── cmd/
│   ├── consumer/   # Consumer service implementation
│   └── generator/  # Generator service implementation
├── docker/         # Dockerfiles
├── k8s/            # Kubernetes manifests
├── pkg/
│   ├── nats/       # NATS client implementation
│   └── redis/      # Redis client implementation
└── web/            # Web UI implementation
```

### Building Changes

1. After modifying the Go code:
```bash
docker build -t generator:latest -f docker/Dockerfile.generator .
docker build -t consumer:latest -f docker/Dockerfile.consumer .
kubectl rollout restart deployment/generator deployment/consumer
```

2. For UI-only changes:
```bash
cd web && npm run build && cd ..
docker build -t generator:latest -f docker/Dockerfile.generator .
kubectl rollout restart deployment/generator
```

## License

MIT License 