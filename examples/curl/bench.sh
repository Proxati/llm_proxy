#!/bin/bash

# This script sends N number of requests at the same time to a LLM API.

# Check if OPENAI_API_KEY is defined
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Error: OPENAI_API_KEY is not set. Please define the environment variable."
    exit 1
fi

PROMPT_API_URL="http://api.openai.com/v1/chat/completions"
PROXY_URL="http://localhost:8080"

# Default values, if the environment variables aren't defined at runtime
SYSTEM_PROMPT="${SYSTEM_PROMPT:-You are a helpful assistant that does meal planning.}"
USER_PROMPT="${USER_PROMPT:-Create a healthy lunch menu}"
GPT_MODEL="${GPT_MODEL:-gpt-4o-mini}"
REQUEST_COUNT="${REQUEST_COUNT:-10}"  # Number of requests to send

# Store the sequence in a variable
seq_var=$(seq 1 $REQUEST_COUNT)

# Array to store PIDs of background processes
pids=()

for i in $seq_var; do
    curl --silent $PROMPT_API_URL \
        -x $PROXY_URL \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${OPENAI_API_KEY}" \
        -d "{\"model\": \"${GPT_MODEL}\", \"messages\": [
                {\"role\": \"system\",
                       \"content\": \"${SYSTEM_PROMPT}\"
                },
                {\"role\": \"user\",
                       \"content\": \"${USER_PROMPT}\"}]}" &
    pids+=($!)  # Add the PID of the last background process to the array
done

# Wait for all background processes to finish
for pid in "${pids[@]}"; do
    wait $pid
done
