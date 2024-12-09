# Step 1: Build the Go binary
FROM golang:1.23-alpine as builder

# Installs dependencies for compilation and UPX
RUN apk update && apk add --no-cache \
    build-base \
    upx \
    git \
    && rm -rf /var/cache/apk/*

# Working directory where the Go code will be copied
WORKDIR /app

# Copies go.mod and go.sum to resolve dependencies
COPY examples/go.mod examples/go.sum ./

RUN go mod tidy

COPY examples/ .

# Compiles the Go binary in production-optimized mode
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/examples/app .

# Step 2: Apply UPX to the binary in a separate intermediate step
FROM alpine:latest as compress-step

# Install UPX
RUN apk add --no-cache upx

# Copy the Go binary from the builder step and apply UPX
COPY --from=builder /bin/examples/app /app/app
RUN upx --best --ultra-brute /app/app

# Step 3: Creating the final image using scratch
FROM scratch

# Create a non-root user and group with meaningful names
USER nobody:nogroup

# Copy the optimized binary from the UPX step
COPY --from=compress-step /app/app /app/app

# Sets the default command to run the binary
CMD ["/app/app"]
