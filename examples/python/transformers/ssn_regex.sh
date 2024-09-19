#!/bin/sh

# This script runs the social security number regex replacer to filter requests from llm_proxy.
# When this is run, it reads from stdin, and will respond with filtered content on stdout in JSON.
# The JSON will contain a response object that has any social security formatted numbers replaced
# with new random numbers. The filtering in this script is done with a simple regex, and so it
# could be easily tricked. This is just an example of how to use the transformers in a simple way.

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
poetry run python "${SCRIPT_DIR}/ssn_regex.py" stdin
