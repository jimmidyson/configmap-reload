FROM oraclelinux:7-slim AS build_base

LABEL maintainer = "Verrazzano developers <verrazzano_ww@oracle.com>"
ENV GOBIN=/usr/bin
ENV GOPATH=/go
RUN set -eux; \
    yum update -y ; \
    yum-config-manager --save --setopt=ol7_ociyum_config.skip_if_unavailable=true ; \
    yum install -y oracle-golang-release-el7 ; \
    yum-config-manager --add-repo http://yum.oracle.com/repo/OracleLinux/OL7/developer/golang113/x86_64 ; \
	yum install -y \
        git \
        gcc \
        make \
        golang-1.13.3-1.el7 \
	; \
    yum clean all ; \
    go version ; \
	rm -rf /var/cache/yum

# Make sure modules are enabled
ENV GO111MODULE=on

# Fetch all the dependencies
COPY . .
# Build code
RUN go mod download ; \
    make out/configmap-reload-linux-amd64

FROM oraclelinux:7-slim

USER 65534

ARG BINARY=configmap-reload-linux-amd64
COPY --from=build_base out/$BINARY /configmap-reload
COPY THIRD_PARTY_LICENSES.txt LICENSE.txt README.md /licenses/

ENTRYPOINT ["/configmap-reload"]
