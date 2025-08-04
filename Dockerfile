# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.24.5-alpine3.22 AS builder
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.* ./

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on go build -a -o generator main.go

# Use alpine tiny images as a base
FROM alpine:3.22.1

ENV USER_UID=2001 \
    USER_NAME=generator \
    GROUP_NAME=generator

WORKDIR /
COPY --from=builder --chown=${USER_UID} /workspace/generator .

USER ${USER_UID}

ENTRYPOINT ["/generator"]
