#!/bin/bash

ollama serve &
sleep 5
# Pull the desired model
echo "Pulling all-minilm model..."
ollama pull all-minilm

# Keep the container running
wait
