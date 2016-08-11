FROM scratch

COPY out/configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
