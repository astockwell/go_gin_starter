# syntax=docker/dockerfile:1

# USAGE:
# Use `make docker-build` (so GO_VERSION is provided automatically from go.mod)

ARG GO_VERSION=1.24
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder

WORKDIR /app

COPY go.mod go.sum Makefile ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0

RUN BIN_DIR=/out BINARY_NAME=go-gin-starter GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder --chown=65532:65532 /out/go-gin-starter /app/go-gin-starter
COPY --from=builder --chown=65532:65532 /app/config.toml /app/config.toml
COPY --from=builder --chown=65532:65532 /app/templates /app/templates
COPY --from=builder --chown=65532:65532 /app/public /app/public

# 65532 maps to the "nonroot" user in common distroless images and is safe in scratch
# (it avoids using 0 while still aligning with container hardening guidance).
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/go-gin-starter"]
