# Build Stage
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application's source code
COPY . .

RUN mkdir -p /var/logs/app

# Build the Go application
RUN go build -o go_with_me cmd/main.go

# Stage 2: Run the application
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/go_with_me .

# Copy any necessary configuration files
COPY .env .

COPY --from=builder /var/logs/app /var/logs/app

# Expose the application port (adjust if necessary)
EXPOSE 8080

# Run the application
CMD ["./go_with_me"]
