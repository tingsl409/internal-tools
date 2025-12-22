# Kubernetes Deployment Charts

This repository contains Helm charts for deploying a multi-component Kubernetes application stack including ingress, sample applications, and internal tools with monitoring capabilities.

## Overview

The deployment consists of three main Helm charts:

- **ingress** - Nginx-based ingress controller with routing to various services
- **sample-app** - Sample application deployment (Nginx-based)
- **internal-tools** - Internal tooling with monitoring, alerting, and custom security policies

## Charts Structure

### 1. Ingress Chart (`charts/ingress/`)

**Purpose**: Provides ingress routing and load balancing for the application stack.

**Key Features**:
- Nginx-based ingress controller
- Custom upstream configurations for ArgoCD, Grafana, Alertmanager, Prometheus, and internal tools
- Node selector and tolerations for dedicated ingress worker nodes
- Horizontal Pod Autoscaling (HPA) support

**Configuration**:
- Image: `registry.domain.me/nginx:1.0.0`
- SSL certificates must be built into the image at:
  - `/opt/bitnami/nginx/conf/server.crt`
  - `/opt/bitnami/nginx/conf/server.key`
- Deploys on nodes with `worker: ingress` label
- Routes traffic to monitoring stack and internal services

### 2. Sample App Chart (`charts/sample-app/`)

**Purpose**: Sample application deployment for testing the network policies.

**Key Features**:
- Nginx-based application (version 1.24)

### 3. Internal Tools Chart (`charts/internal-tools/`)

**Purpose**: Internal tooling platform with comprehensive monitoring and security features.

**Key Features**:
- Custom application deployment (`registry.domain.me/internal-tools`)
- Prometheus monitoring integration
- Grafana dashboard configurations
- Alertmanager integration
- Custom Resource Definitions (CRDs) for security policies
- RBAC configurations

**Monitoring Components**:
- **Prometheus Rules**: Custom alerting rules for Kubernetes API server monitoring
- **ServiceMonitor**: Prometheus service discovery configuration
- **Grafana Dashboard**: Pre-configured dashboards for application metrics
- **Alertmanager Config**: Alert routing and notification configurations

## Deployment

### Prerequisites

- Kubernetes cluster (1.16+)
- Helm 3.x
- Prometheus Operator (for monitoring features)

### Installation

1. **Deploy Ingress Controller**:
```bash
helm install ingress ./charts/ingress -n ingress-system --create-namespace
```

2. **Deploy Sample Application**:
```bash
helm install sample-app ./charts/sample-app -n sample-app --create-namespace
```

3. **Deploy Internal Tools**:
```bash
helm install internal-tools ./charts/internal-tools -n dev-internal-tools --create-namespace
```

### Configuration

Each chart includes comprehensive `values.yaml` files with configurable options:

- **Resource limits and requests**
- **Replica counts and autoscaling**
- **Image repositories and tags**
- **Service configurations**
- **Security contexts and RBAC**
- **Monitoring and alerting settings**

## Monitoring Stack Integration

The internal-tools chart integrates with the Prometheus monitoring stack:

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **Alertmanager**: Alert routing and notifications
- **ServiceMonitor**: Automatic service discovery

## Development

### Chart Structure
Each chart follows Helm best practices:
- Templated configurations with `_helpers.tpl`
- Comprehensive values files
- Proper labeling and annotations
