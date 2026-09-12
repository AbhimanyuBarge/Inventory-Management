# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# Final runtime image
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server /app/server

EXPOSE 8081

ENV PORT=8081
ENV GIN_MODE=release

CMD ["/app/server"]
