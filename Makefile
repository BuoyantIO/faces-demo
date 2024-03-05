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

FACES_GUI_VERSION=0.8.0
FACES_SERVICE_VERSION=0.8.0

BASEDIR=k8s/01-base

help:
	@echo "'make images' will build the Docker images for the Faces GUI"
	@echo "(version $(FACES_GUI_VERSION)) and the Faces service (version $(FACES_SERVICE_VERSION))."
	@echo "Make sure you update the versions in the Makefile when you change"
	@echo "the images, or Kubernetes might get confused. You will need to"
	@echo "have the DEV_REGISTRY variable set when doing this."
	@echo ""
	@echo "'VERSION=... make chart' will package up the Helm chart into"
	@echo "'faces-chart-$$VERSION.tgz'. You must set VERSION in order to use
	@echo "this target."
	@echo ""
	@echo "'HELM_REGISTRY=... VERSION=... make push-chart' will push the chart"
	@echo "to the given HELM_REGISTRY. You must set both HELM_REGISTRY and VERSION"
	@echo "in order to use this target."
	@echo ""
	@echo "'make yaml' will build $(BASEDIR)/faces-gui.yaml and"
	@echo "$(BASEDIR)/faces.yaml, suitable for feeding to kubectl apply."
	@echo "These will create Kubernetes Services and Deployments for the"
	@echo "Faces GUI and its microservices. No namespace is stamped into"
	@echo "the YAML, so you can use whatever namespace you like when you"
	@echo "apply it."
	@echo ""
	@echo "'make deploy' will build and apply the k8s YAML into the faces"
	@echo "namespace. This should be safe to do repeatedly."
	@echo ""
	@echo "You can also 'make clean' to remove all the Docker-image stuff,"
	@echo "or 'make clobber' to smite everything and completely start over."
.PHONY: help

registry-check:
	@if [ -z "$(DEV_REGISTRY)" ]; then \
		echo "DEV_REGISTRY must be set (e.g. DEV_REGISTRY=docker.io/myregistry)" >&2 ;\
		exit 1; \
	fi
.PHONY: registry-check

clean:
	rm -rf oci
	rm -rf faces-chart-*
.PHONY: clean

clobber: clean
	rm -f $(BASEDIR)/faces.yaml $(BASEDIR)/faces-gui.yaml
.PHONY: clobber

oci: registry-check
	mkdir -p oci

oci/python.img: | oci
	crane pull python:3.10.7 oci/python.img
.PRECIOUS: oci/python.img

oci/faces-gui.layer: oci src/faces-gui/data/index.html src/faces-gui/start
	ocibuild layer dir --prefix application src/faces-gui > oci/faces-gui.layer

# oci/faces-gui.img: oci/python.img oci/faces-gui.layer
# 	ocibuild image build \
# 		--base oci/python.img \
# 		--config.Entrypoint /application/start \
# 		--tag $(DEV_REGISTRY)/faces-gui:$(FACES_GUI_VERSION) \
# 		oci/faces-gui.layer \
# 		> oci/faces-gui.img
# 	docker load -i oci/faces-gui.img
# 	docker push $(DEV_REGISTRY)/faces-gui:$(FACES_GUI_VERSION)

PYTHON_LAYERS = \
	oci/urllib3-1.26.12-py2.py3-none-any.layer \
	oci/idna-3.4-py3-none-any.layer \
	oci/charset_normalizer-2.1.1-py3-none-any.layer \
	oci/certifi-2022.9.24-py3-none-any.layer \
	oci/requests-2.28.1-py3-none-any.layer

python-layers: $(PYTHON_LAYERS)

oci/%.whl: | oci
	ocibuild python getwheel $(patsubst oci/%,%,$(patsubst %.layer,%.whl,$@)) > $(patsubst %.layer,%.whl,$@)

oci/python-platform.yaml: oci/python.img
	ocibuild python inspect --imagefile=oci/python.img > oci/python-platform.yaml
.PRECIOUS: oci/python.img

oci/%.layer: oci/%.whl oci/python-platform.yaml
	ocibuild layer wheel --platform-file oci/python-platform.yaml oci/$*.whl > oci/$*.layer

oci/squashed-python.layer: $(PYTHON_LAYERS)
	ocibuild layer squash $(PYTHON_LAYERS) > oci/squashed-python.layer

oci/faces-service.layer: oci src/faces-service/server.py
	ocibuild layer dir --prefix faces-service src/faces-service > oci/faces-service.layer

oci/faces-gui.img: oci/python.img oci/faces-service.layer oci/squashed-python.layer oci/faces-gui.layer
	ocibuild image build \
		--base oci/python.img \
		--config.Entrypoint /faces-service/server.py \
		--config.Env.append FACES_SERVICE=gui \
		--tag $(DEV_REGISTRY)/faces-gui:$(FACES_GUI_VERSION) \
		oci/squashed-python.layer oci/faces-service.layer oci/faces-gui.layer \
		> oci/faces-gui.img
	docker load -i oci/faces-gui.img
	docker push $(DEV_REGISTRY)/faces-gui:$(FACES_SERVICE_VERSION)

oci/faces-service.img: oci/python.img oci/faces-service.layer oci/squashed-python.layer
	ocibuild image build \
		--base oci/python.img \
		--config.Entrypoint /faces-service/server.py \
		--tag $(DEV_REGISTRY)/faces-service:$(FACES_SERVICE_VERSION) \
		oci/squashed-python.layer oci/faces-service.layer \
		> oci/faces-service.img
	docker load -i oci/faces-service.img
	docker push $(DEV_REGISTRY)/faces-service:$(FACES_SERVICE_VERSION)

# This is just an alias
images: oci/faces-gui.img oci/faces-service.img

$(BASEDIR)/faces-gui.yaml: src/templates/faces-gui.yaml.in FORCE
	sed -e "s%DEV_REGISTRY%$(DEV_REGISTRY)%" \
	    -e "s%FACES_GUI_VERSION%$(FACES_GUI_VERSION)%" \
		< src/templates/faces-gui.yaml.in > $(BASEDIR)/faces-gui.yaml

$(BASEDIR)/faces.yaml: src/templates/faces.yaml.in FORCE
	sed -e "s%DEV_REGISTRY%$(DEV_REGISTRY)%" \
	    -e "s%FACES_SERVICE_VERSION%$(FACES_SERVICE_VERSION)%" \
		< src/templates/faces.yaml.in > $(BASEDIR)/faces.yaml

# This is just an alias
yaml: $(BASEDIR)/faces-gui.yaml $(BASEDIR)/faces.yaml

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

deploy: images $(BASEDIR)/faces-gui.yaml $(BASEDIR)/faces.yaml
	$(MAKE) reset
	$(MAKE) apply

apply:
	kubectl create namespace faces || true
	linkerd inject $(BASEDIR) | kubectl apply -n faces -f -
	@echo "You should now be able to open http://localhost/faces/ in your browser."

reset: FORCE
	kubectl delete ns faces || true

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
