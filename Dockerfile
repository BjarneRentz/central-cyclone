# Build Go binary
FROM mirror.gcr.io/library/golang AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o central-cyclon main.go

# Install tools (git, Node.js 22.x, npm, cdxgen)
FROM mirror.gcr.io/library/node:22-alpine
RUN apk add --no-cache git
RUN npm install -g @cyclonedx/cdxgen
COPY --from=builder /app/central-cyclon /app/central-cyclon


ENTRYPOINT ["/app/central-cyclon"]