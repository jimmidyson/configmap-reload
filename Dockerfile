ARG BASEIMAGE=busybox
FROM $BASEIMAGE

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload

ENTRYPOINT ["/configmap-reload"]
