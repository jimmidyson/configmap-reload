# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian11:nonroot

FROM --platform=${BUILDPLATFORM} golang:1.24.4@sha256:01f861b85f25cd045191578c87ed057b92bd185f2226bbf375c7510ebcb8abf4 AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
