# syntax=docker/dockerfile:1

# ---------- Build API ----------
FROM golang:1.23-alpine AS api-build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---------- Runtime ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app

# Binaire goose (prebuilt, évite de dépendre de la version Go de goose)
ARG TARGETARCH
RUN case "$TARGETARCH" in \
      arm64) GOOSE_ARCH=arm64 ;; \
      *)     GOOSE_ARCH=x86_64 ;; \
    esac && \
    wget -q "https://github.com/pressly/goose/releases/download/v3.27.3/goose_linux_${GOOSE_ARCH}" -O /app/goose && \
    chmod +x /app/goose

COPY --from=api-build /out/api /app/api
COPY db/migrations /app/db/migrations
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
