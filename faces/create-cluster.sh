#!/bin/env bash
clear

# Create a K3d cluster to run the Faces application.
CLUSTER=${CLUSTER:-faces}
# echo "CLUSTER is $CLUSTER"

SETUP=${SETUP:-setup-cluster.sh}
# echo "SETUP is $SETUP"

# Ditch any old cluster...
k3d cluster delete $CLUSTER &>/dev/null

#@SHOW

# Expose ports 80 and 443 to the local host, so that our ingress can work.
# Also, don't install traefik, since we'll be putting Linkerd on instead.
k3d cluster create $CLUSTER \
    -p "80:80@loadbalancer" -p "443:443@loadbalancer" \
    --k3s-arg '--no-deploy=traefik@server:*;agents:*'

#@wait
#@HIDE
$SHELL $SETUP
