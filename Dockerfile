# Copyright (C) 2022, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

FROM ghcr.io/oracle/oraclelinux:8-slim

RUN microdnf upgrade -y \
    && microdnf clean all

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload

COPY LICENSE.txt THIRD_PARTY_LICENSES.txt SECURITY.md README.md /licenses/

ENTRYPOINT ["/configmap-reload"]
