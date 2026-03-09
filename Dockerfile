# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates build-base

# Set the working directory
WORKDIR /app

# Enable Go modules
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -a -o /go/bin/origadmin .

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /go/bin/origadmin .

# Copy any other necessary files (configs, templates, etc.)
# COPY --from=builder /app/configs ./configs
# COPY --from=builder /app/templates ./templates

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
ENTRYPOINT ["/app/origadmin"]

# Default command arguments
CMD ["--help"]
