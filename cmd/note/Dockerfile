FROM golang:1.25.6-alpine3.23 as base

FROM base AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 \
  go build \
  -ldflags "-s -w -extldflags '-static'" \
  -o note \
  ./cmd/note/main.go

FROM base AS final
WORKDIR /app
COPY --from=builder /app/note .
EXPOSE 8080
ENTRYPOINT ["/app/note"]
