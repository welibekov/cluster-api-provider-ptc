#!/bin/bash

_remove() {
  # Core CAPI resources
  kubectl patch cluster ptc-cluster-demo -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl patch machine ptc-control-plane-node-1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl patch machine ptc-worker-node1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  
  # PTC Infrastructure resources
  kubectl patch ptccluster ptc-cluster-demo -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl patch ptcmachine ptc-control-plane-node-1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl patch ptcmachine ptc-worker-node-1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
}

_ready() {
  kubectl patch cluster ptc-cluster-demo -n default --type=merge --subresource=status -p '{"status":{"infrastructureReady":true}}'
}

_unblock() {
  kubectl get ptcmachines -n default -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | xargs -I {} kubectl patch ptcmachine {} -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl get ptcclusters -n default -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | xargs -I {} kubectl patch ptccluster {} -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl get machines -n default -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | xargs -I {} kubectl patch machine {} -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl get clusters -n default -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | xargs -I {} kubectl patch cluster {} -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl get ipaddressclaims -n default -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' | xargs -I {} kubectl patch ipaddressclaim {} -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  kubectl get ptcmachines,ptcclusters,machines,clusters,ipaddressclaims,ipaddresses -n default
}

main() {
  case "$1" in
    ready)
      _ready
    ;;
    remove)
      _remove
    ;;
    unblock)
      _unblock
    ;;
   *)
     echo "ready|remove|unblock"
     exit 1
    ;;
 esac
}

main "$@"
