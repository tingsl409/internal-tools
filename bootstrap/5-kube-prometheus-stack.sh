#!/bin/bash

## Install kube-prometheus-stack
## retrieve grafana admin password from: kubectl get -n monitoring secret kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d
## INFO: https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack
#
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
kubectl create namespace monitoring
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack --namespace monitoring

## Create the slack webhook for the alert notifications
## INFO: https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks/
#
kubectl create secret generic slack -n dev-internal-tools --from-literal=webhook-url=$(cat ~/.slack/hook)