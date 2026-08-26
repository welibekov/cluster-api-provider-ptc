#!/bin/bash

INSTALL_FORCE=${INSTALL_FORCE:-false}

_install_kubectl() {
  command -v kubectl 2>&1 && return 

  # Download the latest stable binary
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  
  # Make it executable and move to PATH
  chmod +x ./kubectl
  sudo mv ./kubectl /usr/local/bin/kubectl
  
  # Verify installation
  kubectl version --client
}

_install_kind() {
  command -v kind 2>&1 && return

  # Download the latest binary
  curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.22.0/kind-linux-amd64
  
  # Make it executable and move to PATH
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
  
  # Verify installation
  kind version
}

_install_clusterctl() {
  command -v clusterctl 2>&1 && return

  # Download the latest CAPI CLI binary
  curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.9.5/clusterctl-linux-amd64 -o clusterctl
  
  # Make it executable and move to PATH
  chmod +x ./clusterctl
  sudo mv ./clusterctl /usr/local/bin/clusterctl
  
  # Verify installation
  clusterctl version
}

_verify() {
  kubectl version --client && kind version && clusterctl version
}

main() {
  _install_kubectl && \
  _install_kind && \
  _install_clusterctl && \
  _verify
}

main "$@"
