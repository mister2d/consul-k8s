#!/bin/bash

set -eou pipefail

GOSSIP_SECRET_NAME=consul-gossip-key
GOSSIP_KEY=$(vault kv get -field=key kv-v2/consul/gossip/encryption)

NODE_TOKEN=$(vault kv get -field=key kv-v2/consul/token/mgmt)
TOKEN_SECRET_NAME=consul-bootstrap-token

CA_SECRET_NAME=consul-ca-cert
CONSUL_CA_CERT=$(
        curl -kLs https://consul.service.consul:8501/v1/connect/ca/roots | \
        jq -r '.Roots[0].RootCert' \
        )

# Create Kubernetes secret for Consul gossip key
kubectl delete secret \
        --namespace ${NAMESPACE:-default} \
        ${GOSSIP_SECRET_NAME} \
        --ignore-not-found && \
kubectl create secret \
        --namespace ${NAMESPACE:-default} \
        generic ${GOSSIP_SECRET_NAME} \
        --from-literal="key=${GOSSIP_KEY}"

# Create Kubernetes secret for Consul CA
kubectl delete secret \
        --namespace ${NAMESPACE:-default} \
        ${CA_SECRET_NAME} \
        --ignore-not-found && \
kubectl create secret \
        --namespace ${NAMESPACE:-default} \
        generic ${CA_SECRET_NAME} \
        --from-literal="ca.crt=${CONSUL_CA_CERT}"

# Create Kubernetes secret for Consul management token
kubectl delete secret \
        --namespace ${NAMESPACE:-default} \
        ${TOKEN_SECRET_NAME} \
        --ignore-not-found && \
kubectl create secret \
        --namespace ${NAMESPACE:-default} \
        generic ${TOKEN_SECRET_NAME} \
        --from-literal="token=${NODE_TOKEN}"

# Deploy Consul Helm chart
helm install --values client.yaml \
             --namespace ${NAMESPACE:-default} \
             consul-client \
             hashicorp/consul

