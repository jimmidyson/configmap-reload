# Copyright (C) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

FROM ghcr.io/oracle/oraclelinux:7-slim

RUN yum update -y \
    && yum clean all \
    && rm -rf /var/cache/yum \
    && yum install -y nss-3.67.0-4.el7_9.x86_64 nss-tools-3.67.0-4.el7_9.x86_64 nss-sysinit-3.67.0-4.el7_9.x86_64

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload

COPY LICENSE.txt THIRD_PARTY_LICENSES.txt SECURITY.md README.md /licenses/

ENTRYPOINT ["/configmap-reload"]
