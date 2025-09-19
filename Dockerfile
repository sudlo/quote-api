# ---- Stage 1: The Builder ----
# Use the official Go image to build our application
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go app. CGO_ENABLED=0 creates a static binary.
# -o /app/server builds the output file named 'server' in the /app directory.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# ---- Stage 2: The Final Image ----
# Use a minimal 'scratch' image which is completely empty. It's super secure!
FROM scratch

# Set the working directory
WORKDIR /app

# Copy ONLY the compiled binary from the 'builder' stage
COPY --from=builder /app/server .

# Expose port 8080 to the outside world
EXPOSE 8080

# The command to run when the container starts
ENTRYPOINT ["/app/server"]
