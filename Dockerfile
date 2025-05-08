# -------------------------------------------------
# Stage 1: Build the Go code for Linux in a builder
# -------------------------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# If you're on Apple Silicon and your target is arm64-based, this is enough:
RUN go build -o alt-openai .

# If you need to force a particular architecture, e.g., x86_64:
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alt-openai .

WORKDIR /app
RUN mkdir ./images

RUN rm /app/.dockerignore && rm /app/Dockerfile && rm -rf /app/src && rm /app/go.* && rm -rf /app/.git && rm /app/main.go

# Make sure it's executable
RUN chmod +x /app/alt-openai
RUN chmod +x /app/run_alt-openai.sh

# Optional: run it directly as the entrypoint (no shell needed):
ENTRYPOINT ["/app/run_alt-openai.sh"]
    