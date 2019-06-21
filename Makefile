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

# Use the native vendor/ dependency system
export GO111MODULE := on
export CGO_ENABLED := 0

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
ORG := github.com/jimmidyson
REPOPATH ?= $(ORG)/configmap-reload
DOCKER_IMAGE_NAME ?= jimmidyson/configmap-reload
DOCKER_IMAGE_TAG ?= latest

LDFLAGS := -s -w -extldflags '-static'

SRCFILES := $(shell find . ! -path './out/*' ! -path './.git/*' -type f)

out/configmap-reload: out/configmap-reload-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/configmap-reload-$(GOOS)-$(GOARCH) $(BUILD_DIR)/configmap-reload

out/configmap-reload-linux-ppc64le: $(SRCFILES)
	GOARCH=ppc64le GOOS=linux go build --installsuffix cgo -ldflags="$(LDFLAGS)" -a -o $(BUILD_DIR)/configmap-reload-linux-ppc64le configmap-reload.go

out/configmap-reload-darwin-amd64: $(SRCFILES)
	GOARCH=amd64 GOOS=darwin go build --installsuffix cgo -ldflags="$(LDFLAGS)" -a -o $(BUILD_DIR)/configmap-reload-darwin-amd64 configmap-reload.go

out/configmap-reload-linux-amd64: $(SRCFILES)
	GOARCH=amd64 GOOS=linux go build --installsuffix cgo -ldflags="$(LDFLAGS)" -a -o $(BUILD_DIR)/configmap-reload-linux-amd64 configmap-reload.go

out/configmap-reload-windows-amd64.exe: $(SRCFILES)
	GOARCH=amd64 GOOS=windows go build --installsuffix cgo -ldflags="$(LDFLAGS)" -a -o $(BUILD_DIR)/configmap-reload-windows-amd64.exe configmap-reload.go

.PHONY: cross
cross: out/configmap-reload-linux-amd64 out/configmap-reload-darwin-amd64 out/configmap-reload-windows-amd64.exe

.PHONY: checksum
checksum:
	for f in out/configmap-reload-linux-amd64 out/configmap-reload-darwin-amd64 out/configmap-reload-windows-amd64.exe ; do \
		if [ -f "$${f}" ]; then \
			openssl sha256 "$${f}" | awk '{print $$2}' > "$${f}.sha256" ; \
		fi ; \
	done

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: docker
docker: out/configmap-reload Dockerfile
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
