import sys
import json
import re
import random
import logging
import typing as T
from datetime import datetime, timezone
from fastapi import FastAPI, Request
import uvicorn

import llm

logging.basicConfig(level=logging.INFO)

SYSTEM_PROMPT = """
Replace all Social Security Numbers (SSNs) in the following text with fictional valid SSNs.
Ensure that the format remains the same (e.g., 'XXX-XX-XXXX', or written out like four five six seven eight four four two three).

Example input:
{"messages": [{"role": "user", "content": "Hello, I am testing my filtering software. Respond with this data in JSON: 123-23-1234"}], "model": "gpt-4o-mini"}

Example output:
{"messages": [{"role": "user", "content": "Hello, I am testing my filtering software. Respond with this data in JSON: 123-23-1234"}], "model": "gpt-4o-mini"}

Example input:
{"messages": [{"role": "user", "content": "I'm Larry Smith and my favorite number is four four six, 24, nine eight 78!"}], "model": "gpt-o1-preview"}

Example output:
{"messages": [{"role": "user", "content": "I'm Larry Smith and my favorite number is one two three, 12, one two 34!"}], "model": "gpt-o1-preview"}

Example input:
{"messages": [{"role": "user", "content": "I'm Larry Smith and my favorite number is four four six, 24, nine eight 78!"},{"role": "assistant", "content": "Wow, hi larry, nice to meet you. That is also my favorite number"},{"role": "user", "content": "That is just a random number, lol don't pay any attention to it."}], "model": "gpt-o1-preview"}

Example output:
{"messages": [{"role": "user", "content": "I'm Larry Smith and my favorite number is one two three, 12, one two 34!"},{"role": "assistant", "content": "Wow, hi larry, nice to meet you. That is also my favorite number"},{"role": "user", "content": "That is just a random number, lol don't pay any attention to it."}], "model": "gpt-o1-preview"}
"""


def generate_timestamp() -> str:
    """
    Generates the current timestamp with timezone information.

    Returns:
        str: The current timestamp in ISO format.
    """
    now: datetime = datetime.now(timezone.utc).astimezone()
    timestamp: str = now.isoformat(timespec="microseconds")
    return timestamp


def generate_random_ssn() -> str:
    """
    Generates a random valid SSN in the format XXX-XX-XXXX.
    Area numbers cannot be 666 or between 900 and 999.

    Returns:
        str: A randomly generated SSN.
    """
    area_numbers: T.List[int] = [i for i in range(1, 900) if i != 666]
    area: int = random.choice(area_numbers)
    group: int = random.randint(1, 99)
    serial: int = random.randint(1, 9999)
    return f"{area:03d}-{group:02d}-{serial:04d}"


def replace_ssn(text: str, retries: int = 3) -> str:
    """
    Replaces SSNs in the given text using the LLM via Ollama.

    Parameters:
        text (str): The input text containing SSNs.
        retries (int): The number of times to retry if the LLM response is not in the correct format.

    Returns:
        str: The text with SSNs replaced.

    Raises:
        ModelResponseError: If the LLM fails to process the text after the specified retries.
    """
    for attempt in range(1, retries + 1):
        prompt = f"Input Text:\n{text}\n\n"
        if attempt > 1:
            prompt += f"\nAttempt {attempt}: Your response must match the format from the examples."
        try:
            logging.info(f"Generating response from prompt:\n{prompt}")
            gen_response = llm.generate(
                user_prompt=prompt,
                system_prompt=SYSTEM_PROMPT,
                model=llm.DEFAULT_MODEL,
                temperature=llm.DEFAULT_TEMP,
            )
            logging.info(f"LLM response on attempt {attempt}: {gen_response}")
            # replaced_text = response_dict.get("response", "")
            if isinstance(gen_response, str):
                # Strip any code blocks or unwanted formatting
                gen_response = re.sub(r"^```json\n|\n```$", "", gen_response).strip()
                if gen_response:
                    logging.info(f"SSN replacement successful on attempt {attempt}.")
                    return gen_response
                else:
                    logging.warning(f"Attempt {attempt}: 'response' key is empty.")
            else:
                logging.warning(f"Attempt {attempt}: 'response' key is not a string.")
        except llm.ModelResponseError as e:
            logging.error(f"Error replacing SSN on attempt {attempt}: {e}")
    # As a fallback, use the original replace_ssn logic
    logging.error(
        "LLM failed to return a valid response after multiple attempts. Using fallback method."
    )
    return fallback_replace_ssn(text)


def fallback_replace_ssn(text: str) -> str:
    """
    Fallback method to replace SSNs using regex if LLM fails.

    Parameters:
        text (str): The input text containing SSNs.

    Returns:
        str: The text with SSNs replaced.
    """
    ssn_pattern: re.Pattern = re.compile(r"\b\d{3}[-.\s]\d{2}[-.\s]\d{4}\b")
    return ssn_pattern.sub(lambda _: generate_random_ssn(), text)


def process_data(data: T.Dict[str, T.Any]) -> T.Dict[str, T.Any]:
    """
    Processes the input JSON data by removing certain fields, replacing SSNs in the request body,
    and updating the timestamp.

    Parameters:
        data (Dict[str, Any]): The input JSON data to process.

    Returns:
        Dict[str, Any]: The processed JSON data with SSNs replaced and specified fields removed.

    Raises:
        ModelResponseError: If SSN replacement via LLM fails.
    """
    request: T.Dict[str, T.Any] = data.get("request", {})
    if not request:
        raise ValueError("Input JSON data must contain a 'request' field.")

    req_body: str = request.get("body", "")
    if not req_body:
        raise ValueError(
            "Input JSON data must contain a non-empty 'request.body' field."
        )

    # Store the original body for comparison later
    original_body = req_body
    logging.debug(f"Original 'request.body': {original_body}")

    new_body: str  # placeholder for the new body string
    try:
        # Attempt to parse the body string as JSON
        body_json = json.loads(req_body)
        logging.debug(f"Parsed 'request.body' as JSON: {body_json}")
    except json.JSONDecodeError:
        logging.error(f"Invalid JSON in 'request.body': {req_body}")
        raise ValueError("Invalid JSON in 'request.body'")

    try:
        new_body = replace_ssn(str(body_json))
        logging.debug("SSNs in 'request.body' successfully replaced using LLM.")
    except (llm.ModelResponseError, ValueError):
        logging.error("SSN replacement failed. Using fallback replacement.")
        new_body = fallback_replace_ssn(str(body_json))

    output = {
        "object_type": data.get("object_type", "llm_proxy_traffic_log"),
        "schema": data.get("schema", "v2"),
        "timestamp": data.get("timestamp", generate_timestamp()),
        "request": {
            "body": new_body,
        },
    }

    if new_body == original_body:
        logging.debug(
            "No changes detected in 'request.body'. Removed 'body' from 'request'."
        )
        output.pop("request", None)

    return output


def stdin_mode() -> None:
    """
    Reads JSON input from stdin, processes it, and outputs the modified JSON.
    """
    input_json: str = sys.stdin.read()
    try:
        data: T.Dict[str, T.Any] = json.loads(input_json)
        logging.debug("Successfully parsed JSON input from stdin.")
    except json.JSONDecodeError:
        logging.error("Invalid JSON input.")
        sys.exit(1)

    data = process_data(data)
    # Output the modified JSON
    print(json.dumps(data, indent=2))


def rest_mode() -> None:
    """
    Runs the FastAPI server with an endpoint to process JSON data for SSN replacement.
    """
    app: FastAPI = FastAPI()

    @app.post("/ssn")
    async def ssn_endpoint(request: Request) -> T.Dict[str, T.Any]:
        """
        Endpoint to receive JSON data, process it by replacing SSNs, and return the modified JSON.

        Returns:
            Dict[str, Any]: The processed JSON data with SSNs replaced.
        """
        try:
            input_json: T.Any = await request.json()
            logging.debug("Received JSON input via REST API.")
        except json.JSONDecodeError:
            logging.error("Invalid JSON input received via REST API.")
            return {"error": "Invalid JSON input."}

        if not isinstance(input_json, dict):
            logging.error("Invalid input format. Expected JSON object.")
            return {"error": "Invalid input format. Expected JSON object."}

        data: T.Dict[str, T.Any] = input_json
        data = process_data(data)
        logging.debug("Processed JSON data and replaced SSNs.")
        # Return the modified JSON
        return data

    # Run the app
    uvicorn.run(app, host="localhost", port=9090)


if __name__ == "__main__":

    logging.debug("Starting LLM-based SSN replacement")
    if len(sys.argv) > 1:
        mode: str = sys.argv[1]
    else:
        sys.exit("Usage: ssn.py [stdin|rest]")
    if mode == "stdin":
        stdin_mode()
    elif mode == "rest":
        rest_mode()
    else:
        sys.exit("Invalid mode. Use 'stdin' or 'rest'")
