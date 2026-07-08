# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian12@sha256:cba10d7abd3e203428e86f5b2d7fd5eb7d8987c387864ae4996cf97191b33764

FROM --platform=${BUILDPLATFORM} golang:1.27rc2@sha256:2317c8e806fe884a0d5f3d80d596dcd1369a4c211b7eba8d507718245e9e5831 AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
