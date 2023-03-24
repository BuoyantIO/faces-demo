#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: 2022 Buoyant Inc.
# SPDX-License-Identifier: Apache-2.0
#
# Copyright 2022 Buoyant Inc.
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

set -e
clear

# Make sure that we're in the namespace we expect.
kubectl ns default

# Tell demosh to show commands as they're run.
#@SHOW

# Next up: install Emissary-ingress 3.5.0 as the ingress. This is mostly following
# the quickstart, but we use `yq` to delete the downwardAPI volume and volumeMount
# since Giant Swarm disables that. (This means, in turn, that we can't use Ingress
# resources, which we're not using anyway.)
#
# We install this unmeshed for the moment.

#### EMISSARY_INSTALL_START
EMISSARY_CRDS=https://app.getambassador.io/yaml/emissary/3.5.0/emissary-crds.yaml
EMISSARY_INGRESS=https://app.getambassador.io/yaml/emissary/3.5.0/emissary-emissaryns.yaml

kubectl create namespace emissary && \
curl --proto '=https' --tlsv1.2 -sSfL $EMISSARY_CRDS | kubectl apply -f -
kubectl wait --timeout=90s --for=condition=available deployment emissary-apiext -n emissary-system

curl --proto '=https' --tlsv1.2 -sSfL $EMISSARY_INGRESS \
    | yq 'del(.spec.template.spec.containers.[].volumeMounts)' \
    | yq 'del(.spec.template.spec.volumes)' \
    | kubectl apply -f -

kubectl -n emissary wait --for condition=available --timeout=90s deploy -lproduct=aes
#### EMISSARY_INSTALL_END

#@wait
#@clear

# Finally, configure Emissary for HTTP - not HTTPS! - routing to our cluster.
#### EMISSARY_CONFIGURE_START
kubectl apply -f emissary-yaml
#### EMISSARY_CONFIGURE_END

#@wait
#@clear
# Once that's done, install Faces. We also do this without the mesh at first, so
# no ServiceProfiles, but we'll install its Mappings.

#### FACES_INSTALL_START
kubectl create ns faces

kubectl apply -f k8s/01-base
kubectl -n faces wait --for condition=available --timeout=90s deploy --all

#### FACES_INSTALL_END
