FROM gcr.io/distroless/base

COPY out/configmap-reload /configmap-reload

RUN groupadd users && useradd -Mg users nobody
RUN chown nobody:users configmap-reload
USER nobody

ENTRYPOINT ["/configmap-reload"]
