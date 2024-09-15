import sys
import json
import re
import random
from datetime import datetime, timezone


def generate_timestamp():
    # Create a datetime object for the current time with timezone information
    now = datetime.now(timezone.utc).astimezone()
    # Format the datetime object to the desired string format with timezone offset
    timestamp = now.isoformat(timespec='microseconds')
    return timestamp


def generate_random_ssn():
    # Generate a random valid SSN in the format XXX-XX-XXXX
    # Area numbers cannot be 666 or between 900 and 999
    area_numbers = [i for i in range(1, 900) if i != 666]
    area = random.choice(area_numbers)
    group = random.randint(1, 99)
    serial = random.randint(1, 9999)
    return f"{area:03d}-{group:02d}-{serial:04d}"


def replace_ssn(text):
    # Regular expression for SSN: XXX-XX-XXXX, XXX XX XXXX, or XXX.XX.XXXX
    ssn_pattern = re.compile(r'\b\d{3}[-.\s]\d{2}[-.\s]\d{4}\b')
    return ssn_pattern.sub(lambda _: generate_random_ssn(), text)


def replace_ssn_in_data(data):
    if isinstance(data, dict):
        return {key: replace_ssn_in_data(value) for key, value in data.items()}
    elif isinstance(data, list):
        return [replace_ssn_in_data(item) for item in data]
    elif isinstance(data, str):
        return replace_ssn(data)
    else:
        return data


def main():
    input_json = sys.stdin.read()
    data = json.loads(input_json)

    # Remove specified sub-fields from 'request' if they exist
    if 'request' in data:
        for sub_field in ['url', 'method', 'proto', 'header', 'query']:
            data['request'].pop(sub_field, None)

    # Remove 'connection_stats' and 'response' fields if they exist
    for field in ["connection_stats", "response"]:
        data.pop(field, None)

    # Recursively replace SSNs in the entire data
    data = replace_ssn_in_data(data)

    # Replace the timestamp with the current time including timezone info
    data['timestamp'] = generate_timestamp()

    # Output the modified JSON
    print(json.dumps(data, indent=2))


if __name__ == "__main__":
    main()
