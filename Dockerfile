FROM gcr.io/distroless/base
USER	65534

COPY out/configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
