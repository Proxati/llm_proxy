#!/usr/bin/env python
import os
import logging
from openai import OpenAI, DefaultHttpxClient

__doc__ = """
This is a simple example of how to use the OpenAI Python client library with a MiTM proxy server.
"""

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

client = OpenAI(
    base_url=os.environ.setdefault("OPENAI_BASE_URL", "http://api.openai.com/v1"),
    http_client=DefaultHttpxClient(
        proxy="http://localhost:8080",
    ),
)


def main():
    assistant = client.beta.assistants.create(
        instructions="You are a snarky sysadmin. You give helpful answers to technical questions, but you're also a bit of a jerk.",
        name="The Operator",
        model="gpt-4o",
        temperature=0.8,
    )
    try:
        thread = client.beta.threads.create()

        message = client.beta.threads.messages.create(
            thread_id=thread.id,
            role="user",
            content="Can you help me? How does AI work?",
        )

        run = client.beta.threads.runs.create_and_poll(
            thread_id=thread.id,
            assistant_id=assistant.id,
            instructions="You're in the middle of eating a sandwich when you get this question, but you tolerate a response.",
        )
        logger.info("Run completed with status: %s", run.status)

        if run.status != "completed":
            logger.error("Run failed with error: %s", run.error)
            return

        if run.status == "completed":
            messages = client.beta.threads.messages.list(thread_id=thread.id)

        logger.info("messages: ")
        for message in messages:
            if message.content[0].type != "text":
                logger.error("Unexpected message type: %s", message.content[0].type)
                continue
            logger.info(
                {"role": message.role, "message": message.content[0].text.value}
            )

    finally:
        try:
            client.beta.threads.delete(thread.id)
            client.beta.assistants.delete(assistant.id)
        except Exception as e:
            logger.error("Error during cleanup: %s", e)


if __name__ == "__main__":
    main()
