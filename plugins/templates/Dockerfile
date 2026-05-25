FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /plugin .

FROM gcr.io/distroless/base:latest

COPY --from=builder /plugin /plugin

LABEL ade.plugin.name=templates
LABEL ade.plugin.version=1.0.0
LABEL ade.plugin.http-port=8081
LABEL ade.plugin.grpc-port=50051

EXPOSE 8081 50051

ENTRYPOINT ["/plugin"]
