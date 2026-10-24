FROM golang:1.26.3-alpine3.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 \
  go build \
  -ldflags='-s -w' \
  -o billing \
  ./cmd/billing

FROM alpine:3.24 AS final
WORKDIR /app
COPY --from=builder /app/billing .
EXPOSE 8084 18084 28084
ENTRYPOINT ["./billing"]
