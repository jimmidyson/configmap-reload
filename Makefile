# Copyright 2016 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL := /bin/bash -euo pipefail

# Use the native vendor/ dependency system
export GO111MODULE := on
export CGO_ENABLED := 0

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ORG := github.com/jimmidyson
REPOPATH ?= $(ORG)/configmap-reload
DOCKER_IMAGE_NAME ?= jimmidyson/configmap-reload
DOCKER_IMAGE_TAG ?= latest

LDFLAGS := -s -w -extldflags '-static'

SRCFILES := $(shell find . ! -path './out/*' ! -path './.git/*' -type f)

ALL_ARCH=amd64 arm arm64 ppc64le s390x
ML_PLATFORMS=$(addprefix linux/,$(ALL_ARCH))
ALL_BINARIES ?= $(addprefix out/configmap-reload-, \
									$(addprefix linux-,$(ALL_ARCH)) \
									darwin-amd64 \
									windows-amd64.exe)

DEFAULT_BASEIMAGE_amd64   := busybox
DEFAULT_BASEIMAGE_arm     := armhf/busybox
DEFAULT_BASEIMAGE_arm64   := aarch64/busybox
DEFAULT_BASEIMAGE_ppc64le := ppc64le/busybox
DEFAULT_BASEIMAGE_s390x   := s390x/busybox

BASEIMAGE ?= $(DEFAULT_BASEIMAGE_$(GOARCH))

BINARY=configmap-reload-linux-$(GOARCH)

out/configmap-reload: out/configmap-reload-$(GOOS)-$(GOARCH)
	cp out/configmap-reload-$(GOOS)-$(GOARCH) out/configmap-reload

out/configmap-reload-%: $(SRCFILES)
	GOARCH=$(word 2,$(subst -, ,$(*:.exe=))) GOOS=$(word 1,$(subst -, ,$(*:.exe=))) \
		go build --installsuffix cgo -ldflags="$(LDFLAGS)" -a \
		-o $@ configmap-reload.go

.PHONY: cross
cross: $(ALL_BINARIES)

.PHONY: checksum
checksum:
	for f in $(ALL_BINARIES) ; do \
		if [ -f "$${f}" ]; then \
			openssl sha256 "$${f}" | awk '{print $$2}' > "$${f}.sha256" ; \
		fi ; \
	done

.PHONY: clean
clean:
	rm -rf out

.PHONY: docker
docker: out/configmap-reload-$(GOOS)-$(GOARCH) Dockerfile
	docker build --build-arg BASEIMAGE=$(BASEIMAGE) --build-arg BINARY=$(BINARY) -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)-$(GOARCH) .

manifest-tool:
	curl -fsSL https://github.com/estesp/manifest-tool/releases/download/v1.0.0-rc3/manifest-tool-linux-amd64 > ./manifest-tool
	chmod +x ./manifest-tool

.PHONY: push-%
push-%:
	$(MAKE) GOARCH=$* docker
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)-$*

.PHONY: push
push: manifest-tool $(addprefix push-,$(ALL_ARCH)) manifest-push

comma:= ,
empty:=
space:= $(empty) $(empty)
.PHONY: manifest-push
manifest-push: manifest-tool
	./manifest-tool push from-args --platforms $(subst $(space),$(comma),$(ML_PLATFORMS)) --template $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)-ARCH --target $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
