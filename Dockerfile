# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian11:nonroot

FROM --platform=${BUILDPLATFORM} golang:1.25 AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /kvs-tls-reload main.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/ninech/kvs-tls-reloader"

USER 65534

COPY --from=builder /kvs-tls-reload /kvs-tls-reload

ENTRYPOINT ["/kvs-tls-reload"]
