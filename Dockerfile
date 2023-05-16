ARG BASEIMAGE=busybox:1.34.1
FROM $BASEIMAGE

USER 65534

ARG BINARY=configmap-reload
COPY out/$BINARY /configmap-reload

ENTRYPOINT ["/configmap-reload"]
