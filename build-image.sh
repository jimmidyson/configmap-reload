#!/bin/sh

# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

set -x

BASEIMAGE="container-registry.oracle.com/os/oraclelinux:7.8@sha256:46fc083cf0250ed5260fa6fe822d7d4c139ca1f7fc38e4a17ba662464bd1df4a"

if [ -z "${DOCKER_IMAGE_NAME}" ] ; then
    echo "Environment variable DOCKER_IMAGE_NAME not set"
    exit 1
fi
if [ -z "${DOCKER_IMAGE_TAG}" ] ; then
    echo "Environment variable DOCKER_IMAGE_TAG not set"
    exit 1
fi

make out/configmap-reload-linux-amd64

docker build \
    --build-arg BASEIMAGE="${BASEIMAGE}" \
    --build-arg BINARY="configmap-reload-linux-amd64" \
    -t "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" .
