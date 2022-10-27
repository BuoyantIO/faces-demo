#!/usr/bin/env bash

#### FACES_INSTALL_START
kubectl create ns faces

linkerd inject k8s | kubectl apply -f -
#### FACES_INSTALL_END
