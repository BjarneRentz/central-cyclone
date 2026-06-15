# Build Go binary
FROM mirror.gcr.io/library/golang AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o central-cyclon main.go

# Install tools (git, Node.js 22.x, npm, cdxgen)
FROM mirror.gcr.io/library/node:24-alpine
RUN apk add --no-cache maven ca-certificates
RUN npm install -g @cyclonedx/cdxgen

# Create non-root user
RUN addgroup -S appuser && adduser -S appuser -G appuser -h /home/appuser

# Create workspace directory and set permissions
RUN mkdir -p /home/appuser/.central-cyclone/workfolder/repos && \
    mkdir -p /home/appuser/.central-cyclone/workfolder/sboms && \
    chown -R appuser:appuser /home/appuser

COPY --from=builder /app/central-cyclon /app/central-cyclon

# Make sure appuser can execute the binary
RUN chmod +x /app/central-cyclon

USER appuser

ENV HOME=/home/appuser

ENTRYPOINT ["/app/central-cyclon"]