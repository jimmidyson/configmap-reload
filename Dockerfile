# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian11:nonroot

FROM --platform=${BUILDPLATFORM} golang:1.25.1@sha256:d7098379b7da665ab25b99795465ec320b1ca9d4addb9f77409c4827dc904211 AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
