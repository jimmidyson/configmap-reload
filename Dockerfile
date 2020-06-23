FROM golang:1.14 as builder
COPY . /workspace
WORKDIR /workspace
RUN make out/configmap-reload

FROM alpine
USER 65534
COPY --from=builder /workspace/out/configmap-reload /configmap-reload
ENTRYPOINT ["/configmap-reload"]
