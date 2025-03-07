#!/bin/bash

# Start Ollama in the background
# OLLAMA_DEBUG=0 ollama serve &

OLLAMA_DEBUG=0 ollama serve > /dev/null 2>&1 &

# Wait for Ollama to start
sleep 5

OLLAMA_DEBUG=0 ollama pull llava > /dev/null 2>&1

exit