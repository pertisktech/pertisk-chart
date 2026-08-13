# Build pertisk-chart RPMs on AlmaLinux (el8/el9/el10).
#   docker buildx build --output type=local,dest=dist \
#     --build-arg VERSION=0.1.2 -f packaging/rpm/Dockerfile .
ARG ALMA_VERSION=9
FROM almalinux:${ALMA_VERSION} AS builder

ARG GO_VERSION=1.21.13
ARG VERSION=0.1.2
ARG RELEASE=1

RUN dnf -y install \
        rpm-build rpmdevtools gcc make git tar gzip rsync \
        libcap shadow-utils systemd systemd-rpm-macros \
        ca-certificates \
    && dnf clean all

RUN ARCH="$(uname -m)" \
    && case "$ARCH" in \
         x86_64) GOARCH=amd64 ;; \
         aarch64) GOARCH=arm64 ;; \
         *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;; \
       esac \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" \
         | tar -C /usr/local -xz

ENV PATH=/usr/local/go/bin:$PATH \
    GOPATH=/go \
    GO111MODULE=on \
    CGO_ENABLED=1 \
    GOPROXY=https://proxy.golang.org,direct

WORKDIR /src
COPY . .

RUN chmod +x packaging/rpm/build.sh packaging/pertisk-chart.sh \
        packaging/rpm/preinstall.sh packaging/rpm/postinstall.sh \
        packaging/rpm/preremove.sh \
    && VERSION="$VERSION" RELEASE="$RELEASE" ./packaging/rpm/build.sh \
    && mkdir -p /out \
    && cp /src/dist/*.rpm /out/

FROM scratch
COPY --from=builder /out /
