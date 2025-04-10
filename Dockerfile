# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian11:nonroot

FROM --platform=${BUILDPLATFORM} golang:1.24.2@sha256:18a1f2d1e1d3c49f27c904e9182375169615c65852ace724987929b910195b2c AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
