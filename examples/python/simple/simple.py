#!/usr/bin/env python
import os

from openai import OpenAI, DefaultHttpxClient

__doc__ = """
This is a simple example of how to use the OpenAI Python client library with a MiTM proxy server.
"""

client = OpenAI(
    # max_retries=0,
    base_url=os.environ.setdefault("OPENAI_BASE_URL", "http://api.openai.com/v1"),
    http_client=DefaultHttpxClient(
        proxy="http://localhost:8080",
    ),
)

if __name__ == "__main__":
    # import ipdb; ipdb.set_trace()
    chat_completion = client.chat.completions.create(
        messages=[
            {
                "role": "user",
                "content": "Hello, you are amazing.",
            }
        ],
        model="gpt-4o-mini",
    )
    print(chat_completion)
