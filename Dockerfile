# syntax=docker/dockerfile:1

ARG BASEIMAGE=busybox

ARG GO_VERSION
FROM --platform=${BUILDOS}/${BUILDARCH} golang:${GO_VERSION} as builder

ENV GOOS ${TARGETOS}
ENV GOARCH ${TARGETARCH}
ENV CGO_ENABLED 0

COPY . /src
WORKDIR /src
RUN go build --installsuffix cgo -ldflags="-s -w -extldflags '-static'" -a -o /configmap-reload configmap-reload.go

FROM ${BASEIMAGE}

LABEL org.opencontainers.image.source="https://github.com/jimmidyson/configmap-reload"

USER 65534

COPY --from=builder /configmap-reload /configmap-reload

ENTRYPOINT ["/configmap-reload"]
