# syntax=docker/dockerfile:1

ARG BASEIMAGE=gcr.io/distroless/static-debian11:nonroot

FROM --platform=${BUILDPLATFORM} golang:1.26rc2@sha256:6ea9974028bfaaa0251e42468ec037fe21e49275a874914f3b42f58beabc84bf AS builder

COPY . /src
WORKDIR /src
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
