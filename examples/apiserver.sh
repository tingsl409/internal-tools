#!/bin/bash

## Run in server running the kind cluster
#
if [[ "$1" == "down" ]]; then
    docker exec dev-control-plane mv /etc/kubernetes/manifests/kube-apiserver.yaml /etc/kubernetes/kube-apiserver.yaml
fi

if [[ "$1" == "up" ]]; then
    docker exec dev-control-plane mv /etc/kubernetes/kube-apiserver.yaml /etc/kubernetes/manifests/kube-apiserver.yaml
fi
