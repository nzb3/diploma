#!/bin/bash

ollama serve &
sleep 5
# Pull the desired model
echo "Pulling all-minilm model..."
ollama pull bge-m3:latest

# Keep the container running
wait
