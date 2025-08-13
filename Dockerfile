# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian11:nonroot

FROM --platform=${BUILDPLATFORM} golang:1.24.6@sha256:034848561f95a942e2163d9017e672f0c65403f699336db4529a908af00dfc98 AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
