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

SHELL=bash

help:
	@echo "'make images' will do local builds of all the Docker images,"
	@echo "without pushing them to a registry and therefore without creating"
	@echo "multiarch manifests. This will leave you with images in the local"
	@echo "Docker cache, tagged with version 'latest-{architecture}' (e.g."
	@echo "'latest-arm64' and 'latest-amd64')."
	@echo ""
	@echo "'VERSION=... make chart' will package up the Helm chart into"
	@echo "'faces-chart-$$VERSION.tgz'. You must set VERSION in order to use
	@echo "this target."
	@echo ""
	@echo "'HELM_REGISTRY=... VERSION=... make push-chart' will push the chart"
	@echo "to the given HELM_REGISTRY. You must set both HELM_REGISTRY and VERSION"
	@echo "in order to use this target."
	@echo ""
	@echo "'make proto" will regenerate Go code from protobuf definitions for"
	@echo "the color workload. Requires protoc-gen-go to be installed."
	@echo ""
	@echo "You can also 'make clean' to remove all the Docker-image stuff,"
	@echo "or 'make clobber' to smite everything and completely start over."
.PHONY: help

proto: pkg/faces/color_grpc.pb.go pkg/faces/color.pb.go

pkg/faces/color_grpc.pb.go pkg/faces/color.pb.go: pkg/faces/color.proto
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/faces/color.proto

images:
	goreleaser release --snapshot --clean

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
