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

ALL_BINARIES ?= $(addprefix out/configmap-reload-, \
									$(addprefix linux-,amd64 ppc64le s390x arm64) \
									darwin-amd64 \
									windows-amd64.exe)

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
docker: out/configmap-reload Dockerfile
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
