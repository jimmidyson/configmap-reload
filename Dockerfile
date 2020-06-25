ARG BASEIMAGE=busybox
FROM $BASEIMAGE

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload
COPY THIRD_PARTY_LICENSES.txt /licenses/

ENTRYPOINT ["/configmap-reload"]
