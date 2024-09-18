#!/bin/sh

# This script runs a local LLM (Language Model) via Ollama to filter requests from the llm_proxy.
# When this script is run, it will block while reading from stdin, and will respond on stdout with JSON.
# The JSON will contain a response object with filtered content.
# Unlike the regex replacer, this script uses a local LLM via Ollama, which must be running before this script will work.

# Ensure that Poetry is installed
if ! command -v poetry > /dev/null 2>&1
then
    echo "Poetry could not be found. Please install Poetry first."
    exit 1
fi

# Get the directory of the ssn_llm.py script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Change the working directory to the parent directory of the script
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PARENT_DIR" || { echo "Failed to change directory to ${PARENT_DIR}"; exit 1; }

# Ensure that Ollama daemon is running
if ! poetry run python -c "import ollama; ollama.ps()" > /dev/null 2>&1
then
    echo "Unable to connext to the Ollama daemon."
    echo "Run 'ollama serve' or connect to a remote host with 'ssh -L 11434:localhost:11434 ollama_host'"
    exit 1
fi

# Run the ssn_llm.py script using Poetry
poetry run python "${SCRIPT_DIR}/ssn_llm.py" stdin
