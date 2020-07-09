# Copyright (C) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

ARG BASEIMAGE=busybox
FROM $BASEIMAGE

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload
COPY THIRD_PARTY_LICENSES.txt /licenses/

ENTRYPOINT ["/configmap-reload"]
