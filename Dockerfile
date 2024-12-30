# Build the manager binary
FROM golang:1.21 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY controllers/ controllers/
COPY internal/ internal/
COPY utils/ utils/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use Alpine as base image
FROM alpine:3.19

WORKDIR /

# Install kubectl using curl
RUN apk add --no-cache curl && \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl && \
    apk del curl

# Copy the manager binary from builder stage
COPY --from=builder /workspace/manager .

# Use non-root user for security
RUN adduser -D -u 65532 nonroot && \
    chown nonroot:nonroot /manager

USER 65532:65532

ENTRYPOINT ["/manager"]