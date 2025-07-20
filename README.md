# SMIT Network API

REST API for SMIT network VLAN management built with Go, featuring CI/CD pipeline and Kubernetes deployment.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Requirements](#requirements)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [CI/CD Pipeline](#cicd-pipeline)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)

## Overview

SMIT Network API is a RESTful service for managing VLANs (Virtual Local Area Networks) in a network infrastructure. The API provides CRUD operations for VLAN configurations and includes comprehensive validation, testing, and deployment automation.

## Features

- **RESTful API** for VLAN management
- **JSON file storage** for data persistence
- **Input validation** for all VLAN parameters
- **Health check endpoint** for monitoring
- **Comprehensive test coverage** (>70%)
- **Automated CI/CD pipeline** with GitHub Actions
- **Kubernetes-ready** with deployment manifests
- **Docker containerization**
- **OpenAPI specification** compliance

## Requirements

- Go 1.21 or higher
- Docker (for containerization)
- Kubernetes cluster (for deployment)
- GitHub account (for CI/CD)

## Project Structure

```
smit/
├── .github/
│   └── workflows/
│       └── ci.yaml         # CI/CD pipeline configuration
├── data/
│   └── data.json           # Initial VLAN data
├── kubernetes/
│   ├── deployment.yaml     # Kubernetes deployment manifest
│   └── service.yaml        # Kubernetes service manifest
├── server/
│   └── api/
│       ├── handlers/       # HTTP request handlers
│       │   ├── handlers.go
│       │   └── handlers_test.go
│       ├── models/         # Data models
│       │   ├── vlan.go
│       │   └── models_test.go
│       └── storage/        # Storage layer implementation
│           ├── storage.go
│           └── storage_test.go
├── test/
│   ├── models_test.go      # Model validation tests
│   └── storage_test.go     # Storage layer tests
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
├── main.go                 # Application entry point
├── openapi.yml             # OpenAPI specification
└── README.md
```

## Getting Started

### Local Development

1. **Clone the repository:**
   ```bash
   git clone https://github.com/jargokoster/smit.git
   cd smit
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

   The API will start on `http://localhost:1234`

4. **Run with custom configuration:**
   ```bash
   export SERVER_PORT=8080
   export DATA_FILE_PATH=/path/to/data.json
   go run main.go
   ```

### Docker Build

1. **Build the Docker image:**
   ```bash
   docker build -t smit .
   ```

2. **Run the container:**
   ```bash
   docker run -p 1234:1234 smit
   ```

## API Documentation

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/vlans` | Get all VLANs |
| POST | `/api/v1/vlans` | Create a new VLAN |
| GET | `/api/v1/vlans/{id}` | Get VLAN by ID |
| PUT | `/api/v1/vlans/{id}` | Update VLAN |
| DELETE | `/api/v1/vlans/{id}` | Delete VLAN |
| GET | `/health` | Health check |

### VLAN Model

```json
{
  "id": 100,
  "name": "Production",
  "vlan_id": 100,
  "subnet": "192.168.100.0/24",
  "gateway": "192.168.100.1",
  "status": "active",
  "created_at": "2024-07-15T10:30:00Z",
  "updated_at": "2024-07-15T10:30:00Z"
}
```

### Input Validation

- **name**: 1-255 characters
- **vlan_id**: 1-4094 (valid VLAN range)
- **subnet**: Valid CIDR notation (e.g., 192.168.1.0/24)
- **gateway**: Valid IPv4 address
- **status**: One of: active, inactive, maintenance

### Example Requests

**Create VLAN:**
```bash
curl -X POST http://localhost:1234/api/v1/vlans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Development",
    "vlan_id": 200,
    "subnet": "192.168.200.0/24",
    "gateway": "192.168.200.1",
    "status": "active"
  }'
```

**Get all VLANs:**
```bash
curl http://localhost:1234/api/v1/vlans
```

**Update VLAN:**
```bash
curl -X PUT http://localhost:1234/api/v1/vlans/200 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Development Updated",
    "vlan_id": 200,
    "subnet": "192.168.200.0/24",
    "gateway": "192.168.200.1",
    "status": "maintenance"
  }'
```

## Testing

### Testing Strategy

The project uses a multi-layered testing approach:

1. **Unit Tests**: Test individual components (models, validation)
2. **Integration Tests**: Test API endpoints with mock storage
3. **Storage Tests**: Test data persistence layer

### Run Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -v -coverprofile=coverage -covermode=atomic

# Generate coverage report
go tool cover -html=coverage -o coverage.html

# Check coverage percentage
go tool cover -func=coverage

# Run tests for specific package
go test ./server/api/handlers -v
go test ./server/api/models -v
go test ./server/api/storage -v
```

### Test Coverage

The project maintains >70% code coverage across all packages:

- **handlers**: Tests all HTTP endpoints and error cases
- **models**: Tests input validation thoroughly
- **storage**: Tests CRUD operations and persistence

## CI/CD Pipeline

The GitHub Actions pipeline consists of four stages:

### 1. Test Stage
- Runs all tests
- Generates coverage report
- Fails if coverage < 70%
- Uploads coverage artifacts

### 2. Build Stage
- Compiles Go binary
- Creates Linux-compatible executable
- Uploads binary artifact

### 3. Docker Stage (main branch only)
- Builds Docker image
- Tags with latest and commit SHA
- Pushes to Docker Hub
- Uses build cache for efficiency

### 4. Deploy Stage (main branch only)
- Updates Kubernetes deployment
- Uses new Docker image
- Verifies deployment status
- Requires Kubernetes secrets setup

### Setting up CI/CD

1. **Configure GitHub Secrets:**
   - `DOCKER_USERNAME`: Docker Hub username
   - `DOCKER_PASSWORD`: Docker Hub password
   - `KUBE_CONFIG`: Base64-encoded kubeconfig

2. **Trigger Pipeline:**
   - Push to main/develop branches
   - Create pull request to main

## Kubernetes Deployment

### Prerequisites

1. **Kubernetes cluster** (local or cloud)
2. **kubectl** configured
3. **Docker image** in registry

### Deploy to Kubernetes

1. **Apply deployment:**
   ```bash
   kubectl apply -f kubernetes/deployment.yaml
   ```

2. **Apply service:**
   ```bash
   kubectl apply -f kubernetes/service.yaml
   ```

3. **Verify deployment:**
   ```bash
   kubectl get pods -l app=smit-api
   kubectl get services -l app=smit-api
   ```

### Kubernetes Features

- **Health checks**: Liveness and readiness probes
- **Resource limits**: CPU and memory constraints
- **Scaling**: 3 replicas by default
- **Service types**: LoadBalancer and ClusterIP
- **ConfigMap**: For data file mounting

### Access the Service

**LoadBalancer (cloud environments):**
```bash
kubectl get service smit-api-service
```

**Port forwarding (local testing):**
```bash
kubectl port-forward service/smit-api-service 8080:80
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | API server port | 1234 |
| `DATA_FILE_PATH` | Path to JSON data file | ./data/data.json |

### Data Persistence

The API uses JSON file storage. In production:
- Mount data file as ConfigMap
- Use persistent volumes for production
- Consider database migration for scale

## Security Considerations

1. **Non-root container**: Runs as unprivileged user
2. **Resource limits**: Prevents resource exhaustion
3. **Health checks**: Ensures availability
4. **Input validation**: Prevents invalid data

## Troubleshooting

### Common Issues

1. **Port already in use:**
   ```bash
   export SERVER_PORT=8080
   ```

2. **Data file not found:**
   - Check file path
   - Ensure proper permissions
   - Verify volume mounts in Kubernetes

3. **Coverage below threshold:**
   - Run tests locally first
   - Check for untested code paths
   - Add missing test cases

## Contributing

1. Fork the repository
2. Create feature branch
3. Write tests for new features
4. Ensure coverage >70%
5. Submit pull request

## License

This project is part of a homework assignment for SMIT work interview.

## Contact

- **Author**: Jargo Kõster
- **Email**: jargo@koster.ee