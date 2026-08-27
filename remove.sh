#!/bin/bash

# Core CAPI resources
kubectl patch cluster ptc-cluster-demo -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
kubectl patch machine ptc-control-plane-node-1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
kubectl patch machine ptc-worker-node1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'

# PTC Infrastructure resources
kubectl patch ptccluster ptc-cluster-demo -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
kubectl patch ptcmachine ptc-control-plane-node-1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
kubectl patch ptcmachine ptc-worker-node-1 -n default --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
