# Copyright (C) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

FROM ghcr.io/oracle/oraclelinux:7-slim

RUN yum update -y \
    && yum-config-manager --enable ol7_u8_security_validation \
    && yum install -y openssl \
    && yum clean all \
    && rm -rf /var/cache/yum

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload

COPY LICENSE.txt THIRD_PARTY_LICENSES.txt SECURITY.md README.md /licenses/

ENTRYPOINT ["/configmap-reload"]
