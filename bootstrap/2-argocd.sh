#!/bin/bash

## Retrieve admin password after installation with: kubectl get -n argocd secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
#
kubectl create ns argocd
kubectl create -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/refs/heads/master/manifests/install.yaml

## Patch the installation to enable exec into pod and insecure forward from nginx ingress, need to enable nginx ingress connection upgrade for the exec to work
## INFO:
## - https://rafaelmedeiros94.medium.com/enabling-argocd-web-shell-to-exec-inside-pods-on-amazon-eks-0082d808f6c6
#
kubectl patch configmap argocd-cmd-params-cm -n argocd --type merge -p '{"data":{"server.insecure":"true"}}'
kubectl patch configmap argocd-cm -n argocd --type merge -p '{"data":{"exec.enabled":"true"}}'
kubectl patch clusterrole argocd-server --type='json' -p='[{"op": "add", "path": "/rules/-", "value": {"apiGroups": [""], "resources": ["pods/exec"], "verbs": ["create"]}}]' -n argocd
## Restart argocd-server for the patch to take effect
#
kubectl rollout restart deploy -n argocd argocd-server

## Create several Applications for the demo
## Need to setup github ssh access, INFO:
## - https://dev.to/aditya8raj/setup-github-ssh-keys-for-linux-1hib
#
cat <<EOT | kubectl create -f -
apiVersion: v1
data:
  name: cHVibGlj                                                        # public
  project: ZGVmYXVsdA==                                                 # default
  type: Z2l0                                                            # git
  url: aHR0cHM6Ly9naXRodWIuY29tL3RpbmdzbDQwOS9pbnRlcm5hbC10b29scy5naXQ= # https://github.com/tingsl409/internal-tools.git
kind: Secret
metadata:
  annotations:
    managed-by: argocd.argoproj.io
  labels:
    argocd.argoproj.io/secret-type: repository
  name: pubrepo-github
  namespace: argocd
type: Opaque
EOT

## The app for the internal-tools deployment
#
cat <<EOT | kubectl create -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  labels:
    env: dev
  name: internal-tools
  namespace: argocd
spec:
  destination:
    namespace: dev-internal-tools
    server: https://kubernetes.default.svc
  project: default
  source:
    path: deploy/charts/internal-tools
    repoURL: https://github.com/tingsl409/internal-tools.git
    targetRevision: HEAD
  syncPolicy:
    automated:
      enabled: true
    syncOptions:
    - CreateNamespace=true
EOT

## The ingress to access the frontend services in the cluster
#
cat <<EOT | kubectl create -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  labels:
    env: dev
  name: ingress
  namespace: argocd
spec:
  destination:
    namespace: ingress
    server: https://kubernetes.default.svc
  project: default
  source:
    path: deploy/charts/ingress
    repoURL: https://github.com/tingsl409/internal-tools.git
    targetRevision: HEAD
  syncPolicy:
    automated:
      enabled: true
    syncOptions:
    - CreateNamespace=true
EOT

## The sample apps to test network policies
#
cat <<EOT | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  labels:
    env: dev
  name: nginx-one
  namespace: argocd
spec:
  destination:
    namespace: one
    server: https://kubernetes.default.svc
  project: default
  source:
    path: deploy/charts/sample-app
    repoURL: https://github.com/tingsl409/internal-tools.git
    targetRevision: HEAD
    helm:
      parameters:
        - name: "replicaCount"
          value: "2"
  syncPolicy:
    automated:
      enabled: true
    syncOptions:
    - CreateNamespace=true
EOT

cat <<EOT | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  labels:
    env: dev
  name: nginx-two
  namespace: argocd
spec:
  destination:
    namespace: two
    server: https://kubernetes.default.svc
  project: default
  source:
    path: deploy/charts/sample-app
    repoURL: https://github.com/tingsl409/internal-tools.git
    targetRevision: HEAD
    helm:
      parameters:
        - name: "replicaCount"
          value: "3"
  syncPolicy:
    automated:
      enabled: true
    syncOptions:
    - CreateNamespace=true
EOT

cat <<EOT | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  labels:
    env: dev
  name: nginx-three
  namespace: argocd
spec:
  destination:
    namespace: three
    server: https://kubernetes.default.svc
  project: default
  source:
    path: deploy/charts/sample-app
    repoURL: https://github.com/tingsl409/internal-tools.git
    targetRevision: HEAD
    helm:
      parameters:
        - name: "replicaCount"
          value: "1"
  syncPolicy:
    automated:
      enabled: true
    syncOptions:
    - CreateNamespace=true
EOT