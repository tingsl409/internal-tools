# Kubernetes Internal Tools Platform

A comprehensive Kubernetes platform for internal tooling with custom security policies, monitoring, and GitOps deployment capabilities. This repository provides a complete development environment setup using Kind, along with a Go-based internal tools application that implements custom network security policies and Prometheus monitoring.

## 🏗️ Architecture Overview

This repository consists of four main components:

- **Bootstrap Environment** - Complete Kubernetes development setup with Kind
- **Internal Tools Application** - Go-based Kubernetes controller with custom security policies CRD
- **Helm Charts** - Deployment configurations
- **CI/CD Pipeline** - Automated testing, building, and deployment

## 🚀 Quick Start

### Prerequisites

- Go 1.23+ (for development)

### 1. Bootstrap Kubernetes Environment

Follow the detailed setup instructions in [bootstrap/README.md](bootstrap/README.md) to create the Kind cluster and install all platform components.

### 2. Deploy Internal Tools

The application is deployed via ArgoCD GitOps workflow. For testing purposes, you can also deploy directly with Helm. See [deploy/README.md](deploy/README.md) for detailed deployment instructions.

```bash
# Deploy using Helm (for testing only)
cd deploy
helm install internal-tools ./charts/internal-tools -n dev-internal-tools --create-namespace
helm install ingress ./charts/ingress -n ingress-system --create-namespace
```

### 3. Access Services

**Note**: Add the appropriate IP and domains to your `/etc/hosts` (Linux/Mac) or `C:\Windows\System32\drivers\etc\hosts` (Windows):
```
<server-ip> argocd.domain.me grafana.domain.me internal-tools.domain.me prometheus.domain.me alertmanager.domain.me
```

- **ArgoCD**: `https://argocd.domain.me` (admin password: `kubectl get -n argocd secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d`)
- **Grafana**: `https://grafana.domain.me` (admin password: `kubectl get -n monitoring secret kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d`)
- **Internal Tools**: `https://internal-tools.domain.me`
- **Prometheus**: `https://prometheus.domain.me`
- **Alertmanager**: `https://alertmanager.domain.me`

## 📁 Repository Structure

```
├── bootstrap/          # Kubernetes environment setup scripts
│   ├── 1-kind/        # Kind cluster configuration
│   ├── 2-argocd.sh    # ArgoCD installation
│   ├── 3-github-runner.sh  # GitHub Actions runners
│   ├── 4-metrics-server.sh # Metrics server setup
│   └── 5-kube-prometheus-stack.sh  # Monitoring stack
├── golang/            # Internal tools Go application
│   ├── main.go        # Main application with Kubernetes controller
│   ├── types.go       # CustomDeny CRD definitions
│   ├── Dockerfile     # Multi-stage container build
│   └── main_test.go   # Unit tests
├── deploy/            # Helm charts for deployment
│   └── charts/
│       ├── internal-tools/  # Main application chart
│       ├── ingress/         # Nginx ingress controller
│       └── sample-app/      # Sample application for testing
├── examples/          # Testing scenarios and examples
│   ├── apiserver.sh   # API server availability testing
│   ├── customdeny.yaml     # Network policy example
│   └── unhealthy-deployment.yaml  # Resource scheduling test
└── .github/workflows/ # CI/CD pipeline
    └── ci.yml         # Automated testing and deployment
```

## 🛠️ Internal Tools Application

### Features

The Go-based internal tools application provides:

- **Kubernetes API Monitoring** - Tracks API server availability with Prometheus metrics
- **Deployment Metrics** - Monitors replica counts and availability across all deployments
- **Custom Security Policies** - Implements `CustomDeny` CRD for network traffic control
- **Calico Integration** - Automatically creates Calico NetworkPolicies from CustomDeny resources
- **Health Endpoints** - `/healthz` for readiness and `/metrics` for Prometheus scraping (accessible at `https://internal-tools.domain.me/metrics`)

### Custom Resource Definition (CRD)

The application introduces a `CustomDeny` CRD for simplified network policy management:

```yaml
apiVersion: security.internal.io/v1
kind: CustomDeny
metadata:
  name: deny-cross-namespace
  namespace: target-namespace
spec:
  sourceNamespace: "source-namespace"
  sourceLabels:
    app.kubernetes.io/instance: "nginx-source"
  targetLabels:
    app.kubernetes.io/instance: "nginx-target"
```

This automatically creates corresponding Calico NetworkPolicies to deny traffic between specified pods.

### Prometheus Metrics

The application exposes the following metrics:

- `kubernetes_apiserver_up` - API server connectivity (1=up, 0=down)
- `deployment_spec_replicas` - Desired replica count per deployment
- `deployment_status_replicas_available` - Available replica count per deployment
- `custom_deny_policies_total` - Counter of CustomDeny operations by status

## 🔄 CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/ci.yml`) implements a three-stage pipeline:

### 1. Test Stage
- Runs on self-hosted `poc_runner`
- Executes Go tests

### 2. Build Stage
- Builds Docker image using multi-stage Dockerfile
- Tags with commit SHA for traceability
- Pushes to private registry `registry.domain.me/internal-tools`

### 3. Development Workflow

The platform implements a complete GitOps workflow from code commit to deployment:

```
Developer Push (golang/**) 
  > GitHub Actions Runner (poc_runner)
    > Job 1: Test (Go tests)
    > Job 2: Build (Docker image → registry.domain.me)
    > Job 3: Deploy (Update Helm Chart values → Git Commit & Push)
      > ArgoCD Detection
        > Helm Chart Deployment
          > Image Pull (from registry.domain.me) & Rollout
            > Prometheus Scraping (/metrics endpoint)
              > Alerting & Monitoring
                > Slack Notifications
```

**Workflow Steps:**
1. **Code Push**: Developer pushes changes to `golang/**` directory
2. **CI/CD Trigger**: GitHub Actions workflow activates on self-hosted `poc_runner`
3. **Test Phase**: Go tests execute
4. **Build Phase**: Docker image builds and pushes to `registry.domain.me/internal-tools:${commit-sha}`
5. **Deploy Phase**: Helm chart `values.yaml` updates with new image tag and commits back to repository
6. **GitOps Detection**: ArgoCD monitors repository and detects chart changes
7. **Deployment**: ArgoCD deploys updated Helm chart to Kubernetes cluster
8. **Image Pull**: Cluster pulls new image from private registry triggering deployment rollout
9. **Monitoring**: Prometheus scrapes `/metrics` endpoint from new pods
10. **Alerting**: Monitoring stack sends notifications to Slack channel for deployment status and health alerts

## 📊 Monitoring & Observability

### Prometheus Stack

The platform includes comprehensive monitoring:

- **Prometheus** - Metrics collection and alerting rules
- **Grafana** - Visualization dashboards for application metrics
- **Alertmanager** - Alert routing with Slack integration
- **Custom Alerts** - `KubernetesAPIServerDown` alert for API server monitoring

### Key Dashboards

- Kubernetes cluster overview
- Deployment health and replica status
- Custom security policy metrics
- API server availability trends

## 🧪 Testing & Validation

The `examples/` directory provides comprehensive testing scenarios. See [examples/README.md](examples/README.md) for detailed testing instructions and validation procedures.

## 🚀 Development

### Local Development

```bash
# Clone repository
git clone <repository-url>
cd <repository-name>

# Build and test Go application
cd golang
go mod tidy
go test -v
go build -o internal-tools

# Run locally (requires kubeconfig)
./internal-tools -kubeconfig ~/.kube/config -address :8080
```

### Container Development

```bash
# Build Docker image
cd golang
docker build -t internal-tools:dev .

# Run in container
docker run -p 8080:8080 -v ~/.kube:/root/.kube internal-tools:dev
```

## 📚 Documentation

- [Bootstrap Setup Guide](bootstrap/README.md) - Detailed environment setup instructions
- [Deployment Guide](deploy/README.md) - Helm chart deployment and configuration
- [Testing Examples](examples/README.md) - Comprehensive testing scenarios and validation

## 📄 License

This project is licensed under the terms specified in [LICENSE.md](LICENSE.md).

## 🆘 Troubleshooting

### Common Issues

**API Server Connection Failed**
```bash
# Check cluster status
kubectl cluster-info

# Verify kubeconfig
kubectl config current-context
```

**CustomDeny Policies Not Applied**
```bash
# Check CRD installation
kubectl get crd customdenies.security.internal.io

# Verify controller logs
kubectl logs -n dev-internal-tools deployment/internal-tools
```

**Monitoring Stack Issues**
```bash
# Check Prometheus targets
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090

# Verify ServiceMonitor
kubectl get servicemonitor -n dev-internal-tools
```

### Support

For issues and questions:
1. Check existing GitHub issues
2. Review troubleshooting documentation
3. Create a new issue with detailed reproduction steps
