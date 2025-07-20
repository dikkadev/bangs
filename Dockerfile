FROM --platform=$BUILDPLATFORM node:20 AS frontend-builder
WORKDIR /app

# Copy frontend source code
COPY web/frontend/ ./frontend/

WORKDIR /app/frontend

# Install bun and build frontend
RUN curl -fsSL https://bun.sh/install | bash
RUN $HOME/.bun/bin/bun install
RUN $HOME/.bun/bin/bun run build

# --- Go Builder Stage ---
FROM golang:1.23-alpine AS builder

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy Go source code & modules
COPY go.mod go.sum ./
RUN go mod download
# Copy remaining source files, excluding the frontend directory
COPY . .

# Copy built frontend from the previous stage into the correct location for embedding
COPY --from=frontend-builder /app/frontend/dist ./web/frontend/dist

# Build both the web server and MCP server
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=$VERSION" -o bangs cmd/bangs-server/main.go
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=$VERSION" -o bangs-mcp mcp/main.go

# --- Final Stage ---
FROM alpine:latest

WORKDIR /app

# Copy both compiled Go binaries from the builder stage
COPY --from=builder /app/bangs ./
COPY --from=builder /app/bangs-mcp ./

# Expose the ports the apps listen on
EXPOSE 8080 8081

# Run the binary
# The bangs.yaml file will need to be mounted as a volume in docker-compose
CMD ["./bangs"]
