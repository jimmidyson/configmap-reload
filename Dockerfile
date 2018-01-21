FROM gcr.io/distroless/base

COPY out/configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
