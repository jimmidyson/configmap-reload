#!/bin/sh

set -x

BASEIMAGE="oraclelinux:7.8@sha256:3b8917d800082c4823c4011c3fc4b098e1d50537838e863718516f401621cb93"

if [ -z "${DOCKER_IMAGE_NAME}" ] ; then
    echo "Environment variable DOCKER_IAMGE_NAME not set"
    exit 1
fi
if [ -z "${DOCKER_IMAGE_TAG}" ] ; then
    echo "Environment variable DOCKER_IAMGE_TAG not set"
    exit 1
fi

make out/configmap-reload-linux-amd64

docker build --build-arg BASEIMAGE="${BASEIMAGE}" -t "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" .
