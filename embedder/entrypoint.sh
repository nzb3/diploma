#!/bin/bash

ollama serve &
sleep 5
# Pull the desired model
echo "Pulling mxbai-embed-large model..."
ollama pull mxbai-embed-large

# Keep the container running
wait
