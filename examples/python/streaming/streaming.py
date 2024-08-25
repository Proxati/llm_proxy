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
                "content": "Tell me a short story about proxy servers.",
            }
        ],
        model="gpt-4o-mini",
        stream=True,
        stream_options={"include_usage": True},
        temperature=0.9,
        max_tokens=250,
    )

    full_chunks = []
    for chunk in chat_completion:
        full_chunks.append(chunk)
        if chunk.choices:
            content = chunk.choices[0].delta.content
            if content and content != "None":
                print(content, end='')

    # for chunk in full_chunks:
    #     print(chunk)