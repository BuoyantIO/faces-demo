#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: 2022 Buoyant Inc.
# SPDX-License-Identifier: Apache-2.0
#
# Copyright 2022-2024 Buoyant Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.  You may obtain
# a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

clear

# Make sure that we're in the namespace we expect.
kubectl ns default

# Tell demosh to show commands as they're run.
#@SHOW

#@clear
# Install Linkerd, per the quickstart.
#### LINKERD_INSTALL_START
curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install-edge | sh

linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd check
#### LINKERD_INSTALL_END

linkerd viz install | kubectl apply -f -
linkerd check

#@wait
#@clear
# Next up: install Emissary-ingress 3.9.1 as the ingress.
#
# This is actually cheating quite a bit, by using an unofficial Helm chart
# for the CRDs so that we can disable Emissary's conversion webhook to speed
# up the deployment. We also force every Deployment to one replica to reduce
# the load on k3d.

kubectl create ns emissary
kubectl annotate ns emissary linkerd.io/inject=enabled

helm install emissary-crds \
  oci://registry-1.docker.io/dwflynn/emissary-ingress-crds-chart \
  -n emissary \
  --version 3.9.1 \
  --wait

helm install emissary-ingress \
  oci://ghcr.io/emissary-ingress/emissary-chart \
  -n emissary \
  --version 0.0.0-test \
  --set nameOverride=emissary \
  --set fullnameOverride=emissary \
  --set replicaCount=1

kubectl -n emissary wait --for condition=available --timeout=90s deploy -lproduct=aes

#@wait
#@clear
# Finally, configure Emissary for HTTP - not HTTPS! - routing to our cluster.
kubectl apply -f emissary-yaml

#@wait
#@clear
# Once that's done, install Faces, being sure to inject it into the mesh.
# Install its ServiceProfiles and Mappings too: all of these things are in
# the k8s directory.

kubectl create ns faces
kubectl annotate ns faces linkerd.io/inject=enabled

helm install faces -n faces \
     oci://ghcr.io/buoyantio/faces-chart --version 1.2.0

kubectl rollout status -n faces deploy
kubectl apply -f k8s/01-base

