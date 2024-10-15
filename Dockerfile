FROM golang:1.23-alpine AS builder

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.version=$VERSION" -o bangs .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bangs ./

EXPOSE 8080

CMD ["./bangs"]
