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

SHELL = bash

# Registry for Docker images. Override on the command line:
#   REGISTRY=your.registry.io make push-pubsub
REGISTRY ?= ghcr.io/buoyantio

# Version used for both image tags and Helm chart packaging.
# Defaults to the current git tag; set explicitly for release builds:
#   VERSION=1.2.0 make push-pubsub
#   VERSION=1.2.0 make chart
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

PUBLISHER_IMAGE  := $(REGISTRY)/faces-face-publisher:$(VERSION)
SUBSCRIBER_IMAGE := $(REGISTRY)/faces-face-subscriber:$(VERSION)
ADMIN_IMAGE      := $(REGISTRY)/faces-admin:$(VERSION)
SMILEY_IMAGE     := $(REGISTRY)/faces-smiley:$(VERSION)
COLOR_IMAGE      := $(REGISTRY)/faces-color:$(VERSION)
FACE_IMAGE       := $(REGISTRY)/faces-face:$(VERSION)

# Platforms for multi-arch push (buildx). Override to target a single arch:
#   PLATFORMS=linux/amd64 make push-pubsub
PLATFORMS ?= linux/amd64,linux/arm64

help:
	@echo "=== Admin dashboard ================================================"
	@echo "  make push-admin"
	@echo "      Build and push the faces-admin image (multi-platform)."
	@echo "      REGISTRY=your.registry.io VERSION=x.y.z make push-admin"
	@echo ""
	@echo "=== Classic mode ==================================================="
	@echo "  make images"
	@echo "      Local Docker builds of all images (via goreleaser, no push)."
	@echo "      Images are tagged 'latest-{arch}' in the local cache."
	@echo ""
	@echo "  VERSION=1.0.0 make chart"
	@echo "      Package the Helm chart into faces-chart-\$$VERSION.tgz."
	@echo ""
	@echo "  HELM_REGISTRY=oci://... VERSION=1.0.0 make push-chart"
	@echo "      Push the packaged chart to an OCI Helm registry."
	@echo ""
	@echo "  make proto"
	@echo "      Regenerate Go gRPC code from pkg/color/color.proto."
	@echo "      Requires protoc-gen-go."
	@echo ""
	@echo "=== Pub/Sub mode ===================================================="
	@echo "  make docker-pubsub"
	@echo "      Build publisher + subscriber images locally."
	@echo ""
	@echo "  make push-pubsub"
	@echo "      Build and push both pub/sub images (multi-platform: amd64 + arm64)."
	@echo "      Override with PLATFORMS=linux/amd64 to target a single arch."
	@echo ""
	@echo "  make helm-install-pubsub"
	@echo "      helm upgrade --install in pubsub mode (requires KUBECONFIG)."
	@echo ""
	@echo "  Override registry/version for any pub/sub target:"
	@echo "    REGISTRY=your.registry.io VERSION=1.0.0 make push-pubsub"
	@echo ""
	@echo "=== Maintenance ====================================================="
	@echo "  make clean    Remove built chart tarballs and dist/ directory."
	@echo "  make clobber  Alias for clean."
.PHONY: help

proto: pkg/color/color_grpc.pb.go pkg/color/color.pb.go

pkg/color/color_grpc.pb.go pkg/color/color.pb.go: pkg/color/color.proto
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/color/color.proto

images: .goreleaser.yaml
	goreleaser release --snapshot --clean

.goreleaser.yaml: make-gorel.py gorel.template
	python make-gorel.py < gorel.template > .goreleaser.yaml

clean:
	rm -rf faces-chart-*
	rm -rf dist
.PHONY: clean

clobber: clean
.PHONY: clobber

version-check:
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION must be set (e.g. VERSION=1.0.0-alpha.3)" >&2 ;\
		exit 1; \
	fi
.PHONY: version-check

helm-registry-check:
	@if [ -z "$(HELM_REGISTRY)" ]; then \
		echo "HELM_REGISTRY must be set (e.g. HELM_REGISTRY=oci://ghcr.io/myorganization)" >&2 ;\
		exit 1; \
	fi
.PHONY: helm-registry-check

faces-chart-$(VERSION).tgz: version-check faces-chart
	rm -rf faces-chart-$(VERSION)
	cp -prv faces-chart faces-chart-$(VERSION)
	sed -e "s/%VERSION%/$(VERSION)/" \
		< faces-chart-$(VERSION)/Chart.yaml > faces-chart-$(VERSION)/Chart-fixed.yaml
	mv faces-chart-$(VERSION)/Chart-fixed.yaml faces-chart-$(VERSION)/Chart.yaml
	helm package ./faces-chart-$(VERSION)

push-chart: version-check helm-registry-check faces-chart-$(VERSION).tgz
	if [ -n "$(HELM_REGISTRY)" ]; then \
		helm push faces-chart-$(VERSION).tgz $(HELM_REGISTRY); \
	else \
		echo "HELM_REGISTRY not set, not pushing"; \
	fi

# This is just an alias
chart: faces-chart-$(VERSION).tgz

# Sometimes we have a file-target that we want Make to always try to
# re-generate (such as compiling a Go program; we would like to let
# `go install` decide whether it is up-to-date or not, rather than
# trying to teach Make how to do that).  We could mark it as .PHONY,
# but that tells Make that "this isn't a real file that I expect to
# ever exist", which has a several implications for Make, most of
# which we don't want.  Instead, we can have them *depend* on a .PHONY
# target (which we'll name "FORCE"), so that they are always
# considered out-of-date by Make, but without being .PHONY themselves.
.PHONY: FORCE

# =============================================================================
# Pub/Sub mode
# =============================================================================

.PHONY: docker-pubsub docker-publisher docker-subscriber push-pubsub \
        helm-install-pubsub docker-admin push-admin push-smiley push-color push-face

## Build both pub/sub images locally
docker-pubsub: docker-publisher docker-subscriber

## Build the publisher image only
docker-publisher:
	docker build \
		-f Dockerfiles/Dockerfile.face-publisher \
		-t $(PUBLISHER_IMAGE) \
		.

## Build the subscriber image only
docker-subscriber:
	docker build \
		-f Dockerfiles/Dockerfile.face-subscriber \
		-t $(SUBSCRIBER_IMAGE) \
		.

## Build and push both pub/sub images (multi-platform via buildx)
## Builds directly to the registry — does not update the local Docker cache.
push-pubsub:
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-f Dockerfiles/Dockerfile.face-publisher \
		-t $(PUBLISHER_IMAGE) \
		.
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-f Dockerfiles/Dockerfile.face-subscriber \
		-t $(SUBSCRIBER_IMAGE) \
		.

## Build and push smiley image (includes HTTP PUT handler for admin emoji updates)
push-smiley:
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-f Dockerfiles/Dockerfile.smiley \
		-t $(SMILEY_IMAGE) \
		.

## Build and push color image (includes gRPC UpdateColor for admin color updates)
push-color:
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-f Dockerfiles/Dockerfile.color \
		-t $(COLOR_IMAGE) \
		.

## Build and push face image (classic mode; includes /chaos endpoint for runtime fault injection)
push-face:
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-f Dockerfiles/Dockerfile.face \
		-t $(FACE_IMAGE) \
		.

## Build the admin image locally (single platform)
docker-admin:
	docker build \
		-f Dockerfiles/Dockerfile.faces-admin \
		-t $(ADMIN_IMAGE) \
		.

## Build and push the admin image (multi-platform via buildx)
push-admin:
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		-f Dockerfiles/Dockerfile.faces-admin \
		-t $(ADMIN_IMAGE) \
		.


## helm upgrade --install in pubsub mode (requires KUBECONFIG)
helm-install-pubsub:
	helm upgrade --install faces ./faces-chart \
		--namespace faces \
		--create-namespace \
		--set faceMode=pubsub \
		--set facePublisher.image=$(PUBLISHER_IMAGE) \
		--set faceSubscriber.image=$(SUBSCRIBER_IMAGE)
