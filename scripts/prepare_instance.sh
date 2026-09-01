#!/bin/bash
#
#

_install_hard_prereq() {
  echo "==> installing golang"
  sudo apt update
  sudo apt purge golang-go
  
  wget https://golang.org/dl/go1.22.3.linux-amd64.tar.gz
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
  
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
  source ~/.profile
  go version

  echo "==> installing make"
  sudo apt -y update && sudo apt -y install make

  echo "==> installing docker"
  for pkg in docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc; do sudo apt-get remove $pkg; done
  
  # Add Docker's official GPG key:
  sudo apt-get update
  sudo apt-get install -y ca-certificates curl
  sudo install -m 0755 -d /etc/apt/keyrings
  sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  sudo chmod a+r /etc/apt/keyrings/docker.asc
  
  # Add the repository to Apt sources:
  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
  sudo apt-get update
  
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  
  sudo usermod -aG docker $USER
  newgrp docker
}

_install_tools() {
  echo "==> installing kustomize"
  # Download the binary to current directory
  curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
  
  # Move it to /usr/local/bin so it is globally available
  sudo mv kustomize /usr/local/bin/
}

_install_provider_repo() {
  _install_tools

  echo "==> cloning repository"
  git clone https://github.com/welibekov/cluster-api-provider-ptc
  cd cluster-api-provider-ptc/ || exit 1

  echo "==> installing operate components"
  bash scripts/operate.sh init && \
    bash scripts/build.sh setup
  
  echo "==> create kind cluster"
  kind create cluster --name kind
  
  echo "==> building docker image"
  # Canonical image tag required by clusterctl
  docker build -t localhost/controller:dev .
  kind load docker-image localhost/controller:dev --name kind
  
  echo "==> configure clusterctl provider from github"
  mkdir -p ~/.cluster-api
  
  cat <<'EOF' > ~/.cluster-api/clusterctl.yaml
  providers:
    - name: "ptc"
      url: "https://github.com/welibekov/cluster-api-provider-ptc/releases/v0.1.0/infrastructure-components.yaml"
      type: "InfrastructureProvider"
EOF
  
  echo "==> initializing clusterctl"
  clusterctl init --infrastructure ptc --core cluster-api --bootstrap kubeadm --control-plane kubeadm --ipam in-cluster
}

_install_provider_locally() {
  _install_tools

  echo "==> cloning repository"
  git clone https://github.com/welibekov/cluster-api-provider-ptc
  cd cluster-api-provider-ptc/ || exit 1
  
  echo "==> installing operate components"
  bash scripts/operate.sh init && \
    bash scripts/build.sh setup
  
  echo "==> create kind cluster"
  kind create cluster --name kind
  
  echo "==> building docker image"
  docker build -t localhost/controller:dev .
  kind load docker-image localhost/controller:dev --name kind
  
  echo "==> prepare ptc provider"
  make manifests || exit 1
  mkdir -p ~/.cluster-api/overrides/infrastructure-ptc/v0.1.0
  
  cat <<'EOF' > ~/.cluster-api/clusterctl.yaml
  providers:
    - name: "ptc"
      url: "/home/ubuntu/.cluster-api/overrides/infrastructure-ptc/v0.1.0/infrastructure-components.yaml"
      type: "InfrastructureProvider"
EOF
  
  cat <<'EOF' > ~/.cluster-api/overrides/infrastructure-ptc/v0.1.0/metadata.yaml
  apiVersion: clusterctl.cluster.x-k8s.io/v1alpha3
  kind: Metadata
  releaseSeries:
    - major: 0
      minor: 1
      contract: v1beta1
EOF
  kustomize build config/default > ~/.cluster-api/overrides/infrastructure-ptc/v0.1.0/infrastructure-components.yaml
  sed -i 's|image: controller:latest|image: localhost/controller:dev|g' ~/.cluster-api/overrides/infrastructure-ptc/v0.1.0/infrastructure-components.yaml
  clusterctl init --infrastructure ptc --core cluster-api --bootstrap kubeadm --control-plane kubeadm --ipam in-cluster
}

main() {
  case "$1" in
    init)
      shift
      _install_hard_prereq "$@"
    ;;
    local-install)
      shift
      _install_provider_locally "$@"
    ;;
    repo-install)
      shift
      _install_provider_repo "$@"
    ;;
    *)
      echo "init|local-install|repo-install"
      exit 1
  esac
}

main "$@" 
