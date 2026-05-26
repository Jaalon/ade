# Stage 1 : Build Go
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ade-config ./cmd/ade-config

# Stage 2 : Frontend (placeholder — sera complété dans tâche #005)
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
# Le frontend sera créé dans la tâche #005, le Dockerfile sera
# mis à jour à ce moment pour builder le frontend ici.

# Stage 3 : Image finale
FROM alpine:3.21
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /ade-config /usr/local/bin/ade-config
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
EXPOSE 8080 9090
ENTRYPOINT ["ade-config"]
