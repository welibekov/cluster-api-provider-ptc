#!/usr/bin/env bash

CLUSTER_API_REPO=${CLUSTER_API_REPO:-github.com/welibekov/cluster-api-provider-ptc}
CLUSTER_API_DIRECTORY=${CLUSTER_API_DIRECTORY:-capptc}
CLUSTER_API_DIRECTORY_CREATE=${CLUSTER_API_DIRECTORY_CREATE:-true}

_setup_kubebuilder() {
  # Detect architecture and download binary
  ARCH=$(go env GOARCH)
  echo "Downloading kubebuilder for architecture: $ARCH"
  
  if ! curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/linux/${ARCH}"; then
    echo "Failed to download kubebuilder."
    exit 1
  fi

  # Make binary executable and move to PATH
  chmod +x kubebuilder
  sudo mv kubebuilder /usr/local/bin/

  # Verify installation
  if ! kubebuilder version; then
    echo "Kubebuilder installation failed."
    exit 1
  fi
}

_init_kubebuilder() {
  # Initialize Kubebuilder with the CAPI-compliant domain
  kubebuilder init \
    --domain cluster.x-k8s.io \
    --repo "$CLUSTER_API_REPO" \
    --plugins go/v4
  
  if [ $? -ne 0 ]; then
    echo "Project initialization failed."
    exit 1
  fi
}

_scaffold_kubebuilder() {
  # 1. Create PTCCluster API and Reconciler
  kubebuilder create api \
    --group infrastructure \
    --version v1alpha1 \
    --kind PTCCluster \
    --resource=true \
    --controller=true
  
  if [ $? -ne 0 ]; then
    echo "Failed to create PTCCluster API."
    exit 1
  fi

  # 2. Create PTCMachine API and Reconciler
  kubebuilder create api \
    --group infrastructure \
    --version v1alpha1 \
    --kind PTCMachine \
    --resource=true \
    --controller=true
  
  # 3. Create PTCMachineTemplate 
  kubebuilder create api \
    --group infrastructure \
    --version v1alpha1 \
    --kind PTCMachineTemplate \
    --resource=true \
    --controller=false
      
  if [ $? -ne 0 ]; then
    echo "Failed to create PTCMachine API."
    exit 1
  fi
}

_generate() {
  # 1. Generate code (zz_generated.deepcopy.go)
  make generate
  if [ $? -ne 0 ]; then
    echo "Code generation failed."
    exit 1
  fi

  # 2. Generate CRDs and RBAC manifests in config/
  make manifests
  if [ $? -ne 0 ]; then
    echo "Manifest generation failed."
    exit 1
  fi

  # 3. Test compile the binary
  make build
  if [ $? -ne 0 ]; then
    echo "Build failed."
    exit 1
  fi
}

main() {
  if [[ $CLUSTER_API_DIRECTORY_CREATE = true ]]; then
    [[ -d $CLUSTER_API_DIRECTORY ]] || mkdir -p $CLUSTER_API_DIRECTORY
    cd "$CLUSTER_API_DIRECTORY" || { echo "Failed to create or navigate to project directory $CLUSTER_API_DIRECTORY"; exit 1; }
  fi

  # Check for arguments
  if [ "$#" -ne 1 ]; then
    echo "Usage: $0 {setup|init|scaffold|generate|all|almost-all}"
    exit 1
  fi

  case $1 in
    setup)
      _setup_kubebuilder
      ;;
    init)
      _init_kubebuilder
      ;;
    scaffold)
      _scaffold_kubebuilder
      ;;
    generate)
      _generate
      ;;
    almost-all)
      _init_kubebuilder && \
      _scaffold_kubebuilder && \
      _generate
      ;;
    all)
      _setup_kubebuilder && \
      _init_kubebuilder && \
      _scaffold_kubebuilder && \
      _generate
      ;;
    *)
      echo "Error: Invalid option '$1'"
      echo "Possible choices are: setup, init, scaffold, generate, all, almost-all"
      exit 1
      ;;
  esac
}

main "$@"
