#!/bin/sh

# Ensure that Poetry is installed
if ! command -v poetry > /dev/null 2>&1
then
    echo "Poetry could not be found. Please install Poetry first."
    exit 1
fi

# Get the directory of the ssn.py script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Change the working directory to the parent directory of the script
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PARENT_DIR" || { echo "Failed to change directory to ${PARENT_DIR}"; exit 1; }

# Run the ssn.py script using Poetry
poetry run python "${SCRIPT_DIR}/ssn.py"
