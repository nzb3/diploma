#!/bin/bash

ollama serve &
sleep 5
# Pull the desired model
echo "Pulling model..."
ollama pull gemma3:1b

wait
