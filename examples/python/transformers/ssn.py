import sys
import json
import re
import random
from datetime import datetime, timezone
from fastapi import FastAPI, Request
import uvicorn
from typing import Any, Dict, List


def generate_timestamp() -> str:
    # Create a datetime object for the current time with timezone information
    now: datetime = datetime.now(timezone.utc).astimezone()
    # Format the datetime object to the desired string format with timezone offset
    timestamp: str = now.isoformat(timespec="microseconds")
    return timestamp


def generate_random_ssn() -> str:
    # Generate a random valid SSN in the format XXX-XX-XXXX
    # Area numbers cannot be 666 or between 900 and 999
    area_numbers: List[int] = [i for i in range(1, 900) if i != 666]
    area: int = random.choice(area_numbers)
    group: int = random.randint(1, 99)
    serial: int = random.randint(1, 9999)
    return f"{area:03d}-{group:02d}-{serial:04d}"


def replace_ssn(text: str) -> str:
    # Regular expression for SSN: XXX-XX-XXXX, XXX XX XXXX, or XXX.XX.XXXX
    ssn_pattern: re.Pattern = re.compile(r"\b\d{3}[-.\s]\d{2}[-.\s]\d{4}\b")
    return ssn_pattern.sub(lambda _: generate_random_ssn(), text)


def replace_ssn_in_data(data: Any) -> Any:
    if isinstance(data, dict):
        return {key: replace_ssn_in_data(value) for key, value in data.items()}
    elif isinstance(data, list):
        return [replace_ssn_in_data(item) for item in data]
    elif isinstance(data, str):
        return replace_ssn(data)
    else:
        return data


def process_data(data: Dict[str, Any]) -> Dict[str, Any]:
    original_body: str = ""
    if "request" in data.keys():
        original_body = data["request"].get("body", "")

    # Remove specified sub-fields from 'request' if they exist
    if "request" in data:
        for sub_field in ["url", "method", "proto", "header", "query"]:
            data["request"].pop(sub_field, None)

    # Remove 'connection_stats' and 'response' fields if they exist
    for field in ["connection_stats", "response"]:
        data.pop(field, None)

    # Recursively replace SSNs in the entire data
    data = replace_ssn_in_data(data)

    # Replace the timestamp with the current time including timezone info
    data["timestamp"] = generate_timestamp()

    if "request" in data.keys():
        new_body = data["request"].get("body")
        try:
            if json.loads(new_body) == json.loads(original_body):
                # No changes made, so delete the body from the request to prevent it from being sent
                data["request"].pop("body", None)
        except (json.JSONDecodeError, TypeError):
            # Unable to parse JSON or new_body is None, skip it
            pass

    if data.get("request", {}) == {}:
        data.pop("request", None)

    return data


def stdin_mode() -> None:
    input_json: str = sys.stdin.read()
    try:
        data: Dict[str, Any] = json.loads(input_json)
    except json.JSONDecodeError:
        print("Invalid JSON input")
        sys.exit(1)

    data = process_data(data)
    # Output the modified JSON
    print(json.dumps(data, indent=2))


def rest_mode() -> None:
    app: FastAPI = FastAPI()

    @app.post("/ssn")
    async def ssn_endpoint(request: Request) -> Dict[str, Any]:
        input_json: Any = await request.json()
        data: Dict[str, Any] = input_json
        data = process_data(data)
        # Return the modified JSON
        return data

    # Run the app
    uvicorn.run(app, host="localhost", port=9090)


if __name__ == "__main__":
    if len(sys.argv) > 1:
        mode: str = sys.argv[1]
    else:
        print("Usage: ssn.py [stdin|rest]")
        sys.exit(1)
    if mode == "stdin":
        stdin_mode()
    elif mode == "rest":
        rest_mode()
    else:
        print("Invalid mode. Use 'stdin' or 'rest'")
        sys.exit(1)
