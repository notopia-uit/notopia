FROM golang:1.26.3-alpine3.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 \
  go build \
  -ldflags='-s -w' \
  -o note \
  ./cmd/note

FROM alpine:3.23 AS final
WORKDIR /app
COPY --from=builder /app/note .
EXPOSE 8081 18081 28081
ENTRYPOINT ["./note"]
