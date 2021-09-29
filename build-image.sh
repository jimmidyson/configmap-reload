#!/bin/sh

# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

set -x

BASEIMAGE="container-registry.oracle.com/os/oraclelinux:7.9@sha256:8462ac343340c5dde2343815d3b2cac40332ee0ab15493fbe00bc33f2846b497"

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
