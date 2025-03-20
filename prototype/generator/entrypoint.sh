#!/bin/bash

ollama serve &
sleep 5
# Pull the desired model
echo "Pulling all-minilm model..."
ollama pull llama3

# Keep the container running
wait
