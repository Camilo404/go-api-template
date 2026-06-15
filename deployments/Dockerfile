# syntax=docker/dockerfile:1.7
#
# Multi-stage build for a tiny, static Go binary. The runtime image
# runs as a non-root user on Alpine and contains only the compiled
# binary plus CA certificates.
#
# Build:    docker build -f deployments/Dockerfile -t template-api:latest .
# Run:      docker run --rm -p 8080:8080 template-api:latest

# ---------- build stage ----------
FROM golang:1.25-alpine AS build
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source and build a static binary.
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---------- runtime stage ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=build /out/api /app/api

ENV PORT=8080 \
    APP_ENV=production \
    LOG_LEVEL=info
EXPOSE 8080
ENTRYPOINT ["/app/api"]
