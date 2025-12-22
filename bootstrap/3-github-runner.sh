#!/bin/bash

## Install cert manager dependency
#
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version v1.15.1 --set crds.enabled=true

## Create the PAT to Install Github Action Runner in the path ~/.gittok/github_token
## INFO:
## - https://medium.com/@ashok.kammala/github-actions-runner-on-kubernetes-5c04d5bff175
#
kubectl create ns actions-runner-system
kubectl create secret generic controller-manager -n actions-runner-system --from-literal=github_token=$(cat ~/.gittok/github_token)
## Install Github Action Runner
#
helm repo add actions-runner-controller https://actions-runner-controller.github.io/actions-runner-controller
helm repo update
helm upgrade --install --namespace actions-runner-system --create-namespace --wait actions-runner-controller actions-runner-controller/actions-runner-controller --set syncPeriod=1m

## Create a RunnerDeployment to monitor the repo, mount the hostPath to docker service in the runner
## to get the trusted ca for permission to push and pull from private registry
#
echo -n 'apiVersion: actions.summerwind.dev/v1alpha1
kind: RunnerDeployment
metadata:
  name: k8s-action-runner
  namespace: actions-runner-system
spec:
  replicas: 3
  template:
    spec:
      repository: tingsl409/internal-tools
      volumes:
      - name: ca-crt
        hostPath:
          path: /usr/local/share/ca-certificates/
          type: Directory
      dockerVolumeMounts:
      - name: ca-crt
        mountPath: /etc/docker/certs.d/registry.domain.me/
        readOnly: true
      env:
      - name: ARC_DOCKER_MTU_PROPAGATION
        value: "true"
      labels:
        - poc_runner' | kubectl create -f -