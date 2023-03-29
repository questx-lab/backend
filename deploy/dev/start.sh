#!/bin/bash

helm repo add pingcap https://charts.pingcap.org/

kind create cluster --quiet
helm upgrade --install database pingcap/tidb-cluster --namespace insfrastructure --create-namespace -f ./deploy/helms/database.yaml
helm upgrade --install ${srv} ../deployments/helms/${srv} --namespace ${SERVICES} --create-namespace
