#!/bin/bash

ollama serve &
sleep 5
# Pull the desired model
echo "Pulling model..."
ollama pull gemma3:4b-it-qat


# Keep the container running
wait
