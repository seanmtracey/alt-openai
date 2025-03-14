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
RUN go build -o alt-llava .

# If you need to force a particular architecture, e.g., x86_64:
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alt-llava .

# -------------------------------------------------
# Stage 2: Use the 'ollama/ollama' image
# -------------------------------------------------
FROM ollama/ollama:latest

WORKDIR /app
RUN mkdir ./images

COPY --from=builder /app/alt-llava /app/alt-llava
# RUN ollama serve && sleep 10
COPY --from=builder /app/run_ollama.sh /app/run_ollama.sh
COPY --from=builder /app/run_alt-llava.sh /app/run_alt-llava.sh

# Make sure it's executable
RUN chmod +x /app/alt-llava
RUN chmod +x /app/run_ollama.sh
RUN chmod +x /app/run_alt-llava.sh
RUN /app/run_ollama.sh

# Optional: run it directly as the entrypoint (no shell needed):
ENTRYPOINT ["/app/run_alt-llava.sh"]
    