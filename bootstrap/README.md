# Kubernetes Bootstrap Environment

Bootstrap scripts for setting up a complete Kubernetes development environment with Kind.

## Prerequisites

- Docker, kubectl, helm, kind installed
- GitHub PAT in `~/.gittok/github_token`
- GitHub SSH key setup
- Slack webhook in `~/.slack/hook`
- Domain certificate and key in `~/.domain/domain.me.crt` and `~/.domain/domain.me.key`

## Setup Steps

Execute the scripts in order:

### 1. Create Kind Cluster
```bash
cd 1-kind
## Change apiServerAddress in the below 1-values.yaml to the approriate value in your environment
#
kind create cluster --config 1-values.yaml --name dev
./2-cni-and-registry.sh
kubectl apply -f 3-install-cacrt.yaml
```
**Description**: Creates Kind cluster with 1 control-plane + 3 worker nodes (one tainted for ingress with nodeport), installs Calico CNI with pod network 10.200.0.0/16, sets up default allow-all network policy, and installs trusted CA certificates for private registry access.
**Note**: Add the appropriate IP to your `/etc/hosts` (Linux/Mac) or `C:\Windows\System32\drivers\etc\hosts` (Windows) for the registry URL.

### 2. Install ArgoCD
```bash
cd ..
./2-argocd.sh
```
**Description**: Installs ArgoCD for GitOps deployment, enables exec into pods and insecure nginx ingress forwarding, creates demo applications with ingress access.

**Admin password**: `kubectl get -n argocd secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d`

### 3. Install GitHub Actions Runner
```bash
./3-github-runner.sh
```
**Description**: Installs cert-manager dependency and sets up GitHub Actions self-hosted runners in Kubernetes. Requires GitHub PAT in `~/.gittok/github_token`.

### 4. Install Metrics Server
```bash
./4-metrics-server.sh
```
**Description**: Installs metrics server with insecure kubelet TLS for Kind compatibility.

### 5. Install Prometheus Stack
```bash
./5-kube-prometheus-stack.sh
```
**Description**: Installs kube-prometheus-stack for monitoring and sets up Slack webhook for alert notifications (requires `~/.slack/hook`).

**Grafana password**: `kubectl get -n monitoring secret kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d`

## Cleanup

```bash
kind delete cluster --name dev
```

## References

- [ArgoCD Web Shell Setup](https://rafaelmedeiros94.medium.com/enabling-argocd-web-shell-to-exec-inside-pods-on-amazon-eks-0082d808f6c6)
- [GitHub SSH Keys Setup](https://dev.to/aditya8raj/setup-github-ssh-keys-for-linux-1hib)
- [GitHub Actions Runner on Kubernetes](https://medium.com/@ashok.kammala/github-actions-runner-on-kubernetes-5c04d5bff175)
- [Docker Registry Self-signed Certificates](https://distribution.github.io/distribution/about/insecure/#use-self-signed-certificates)
- [Docker Registry Deployment](https://distribution.github.io/distribution/about/deploying/#get-a-certificate)
- [Containerd Certificate DaemonSet](https://gist.githubusercontent.com/tomconte/25f7db5b419c24db8bc6cac2fa4c2a13/raw/b008ad130b8eacf24df029ed3454d124d3e845ab/containerd-certificate-ds.yaml)
- [Kube Prometheus Stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [Slack Incoming Webhooks](https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks/)
