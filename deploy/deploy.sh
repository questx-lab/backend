#!/bin/bash

# namespace
DATABASE=database
SERVICES=services

# path
BACKEND_PATH=./backend
TIDB_PATH="./deploy/tidb"

require() {
  kind create cluster --quiet
  kubectl create -f https://raw.githubusercontent.com/pingcap/tidb-operator/master/manifests/crd.yaml
}

install_helms() {
  helm repo add pingcap https://charts.pingcap.org/ --force-update

  # for operator(admin)
  helm upgrade --install tidb-operator pingcap/tidb-operator --vers --namespace ${DATABASE}
  helm upgrade --install backend ${BACKEND_PATH} --namespace ${SERVICES} --create-namespace

}

deploy_tidb() {
  kubectl create namespace ${DATABASE}
  # create cluster
  kubectl -n ${DATABASE} apply -f ${TIDB_PATH}/tidb-cluster.yaml

  # create dashboard
  kubectl -n ${DATABASE} apply -f ${TIDB_PATH}/tidb-dashboard.yaml

  # create monitor
  kubectl -n ${DATABASE} apply -f ${TIDB_PATH}/tidb-monitor.yaml
}

forward_port() {
  kubectl port-forward -n ${DATABASE} svc/basic-tidb 14000:4000 >pf14000.out &
}

require
init
install_helms
deploy_tidb
# forward_port
