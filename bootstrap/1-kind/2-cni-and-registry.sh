#!/bin/bash

## Install calico with pod ip network set to 10.200.0.0/16
#
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.2/manifests/operator-crds.yaml
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.2/manifests/tigera-operator.yaml
curl -fsSL https://raw.githubusercontent.com/projectcalico/calico/v3.31.2/manifests/custom-resources.yaml | sed 's/192\.168\.0\.0/10.200.0.0/' | kubectl create -f -

## Setup default allow all in the cluster
#
cat <<EOT | kubectl apply -f -
apiVersion: projectcalico.org/v3
kind: GlobalNetworkPolicy
metadata:
  name: default-allow
spec:
  order: 1000
  selector: all()
  types:
  - Ingress
  - Egress
  ingress:
  - action: Allow
  egress:
  - action: Allow
EOT

## Create the ssl cert and registry with the command below then add the trusted ca to the client machine(s)
## INFO:
## - https://distribution.github.io/distribution/about/insecure/#use-self-signed-certificates
## - https://distribution.github.io/distribution/about/deploying/#get-a-certificate
#
mkdir -p $(realpath ~)/.domain/

openssl req -x509 -newkey rsa:2048 -keyout $(realpath ~)/.domain/domain.me.key -out $(realpath ~)/.domain/domain.me.crt -days 36500 -nodes -subj "/CN=domain.me" -addext "subjectAltName=DNS:domain.me,DNS:*.domain.me"

docker run -d \
  --restart=unless-stopped \
  --name registry \
  -v $(realpath ~)/.domain/:/certs \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.me.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/domain.me.key \
  -p 443:443 \
  registry:3