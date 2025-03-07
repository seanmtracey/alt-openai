#!/bin/sh

# Start Ollama in the background
OLLAMA_DEBUG=0 ollama serve > /dev/null 2>&1 &

# Wait for Ollama to start
sleep 5

/app/alt-llava