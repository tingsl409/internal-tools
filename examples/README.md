# Testing Examples

This folder contains test scenarios and examples for validating the monitoring alerts and security policies deployed by the Helm charts.

## Files Overview

### 1. `apiserver.sh` - API Server Availability Test

**Purpose**: Script to simulate Kubernetes API server downtime for testing monitoring alerts.

**Usage**:
```bash
# Bring API server down (triggers KubernetesAPIServerDown alert)
./apiserver.sh down

# Bring API server back up
./apiserver.sh up
```

**What it does**:
- Moves the API server manifest out of `/etc/kubernetes/manifests/` to stop the API server
- Moves it back to restore service
- Requires access to the Kind cluster control plane container

**Expected Alert**: `KubernetesAPIServerDown` from the internal-tools Prometheus rules

### 2. `customdeny.yaml` - Security Policy Test

**Purpose**: Example CustomDeny resource to test namespace-based network security policies.

**Configuration**:
- **Source**: Namespace `one`, pods with label `app.kubernetes.io/instance: nginx-one`
- **Target**: Namespace `two`, pods with label `app.kubernetes.io/instance: nginx-two`
- **Effect**: Denies network traffic from source to target

**Usage**:
```bash
# Apply the CustomDeny policy
kubectl apply -f customdeny.yaml

# Verify the policy is created
kubectl get customdenies -A

# Test by deploying nginx instances in both namespaces and attempting connection
```

### 3. `unhealthy-deployment.yaml` - Pod Scheduling Test

**Purpose**: Deployment that intentionally creates scheduling conflicts to test cluster behavior.

**Configuration**:
- **Replicas**: 3 (expects more nodes than available)
- **Anti-Affinity**: Requires each pod on a different node
- **Image**: `nginx:alpine`

**Usage**:
```bash
# Deploy the problematic workload
kubectl apply -f unhealthy-deployment.yaml

# Check pod status (some should remain Pending)
kubectl get pods -l app=one-per-node-app

# Check events for scheduling failures
kubectl describe deployment one-per-node-app
```

**Expected Behavior**: Some pods will remain in `Pending` state due to insufficient nodes

## Test Scenarios

### Monitoring Alert Testing

1. **API Server Down Alert**:
   ```bash
   # Trigger alert
   ./apiserver.sh down
   
   # Check Prometheus alerts (should fire within minutes)
   # Check Alertmanager for notifications
   
   # Restore service
   ./apiserver.sh up
   ```

### Security Policy Testing

1. **Network Policy Validation**:
   ```bash
   # Create test namespaces
   kubectl create namespace one
   kubectl create namespace two
   
   # Deploy nginx in both namespaces
   kubectl run nginx-one --image=nginx -n one --labels="app.kubernetes.io/instance=nginx-one"
   kubectl run nginx-two --image=nginx -n two --labels="app.kubernetes.io/instance=nginx-two"
   
   # Apply CustomDeny policy
   kubectl apply -f customdeny.yaml
   
   # Test connectivity (should be blocked)
   kubectl exec -n one nginx-one -- curl nginx-two.two.svc.cluster.local
   ```

### Resource Scheduling Testing

1. **Pod Anti-Affinity**:
   ```bash
   # Deploy the unhealthy deployment
   kubectl apply -f unhealthy-deployment.yaml
   
   # Monitor pod scheduling
   kubectl get pods -l app=one-per-node-app -w
   
   # Check node capacity
   kubectl get nodes
   kubectl describe nodes
   ```

## Prerequisites

- Kubernetes cluster with monitoring stack deployed
- Internal-tools chart installed (for CustomDeny CRD and Prometheus rules)
- Access to cluster control plane (for API server testing)
- kubectl configured for cluster access

## Cleanup

```bash
# Remove test resources
kubectl delete -f unhealthy-deployment.yaml
kubectl delete -f customdeny.yaml
kubectl delete namespace one two

# Ensure API server is running
./apiserver.sh up
```

## Expected Outcomes

- **API Server Test**: Prometheus alert fires and resolves
- **CustomDeny Test**: Network traffic blocked between specified pods
- **Scheduling Test**: Demonstrates resource constraints and anti-affinity rules
