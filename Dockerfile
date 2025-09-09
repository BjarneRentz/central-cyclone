# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o central-cyclon main.go

# Final distroless stage
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/central-cyclon /app/central-cyclon
ENTRYPOINT ["/app/central-cyclon"]
