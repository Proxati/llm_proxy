import json
import logging
import typing as T
import threading  # Imported for thread safety

import ollama  # type: ignore

# Configure logging
logging.basicConfig(level=logging.INFO)

# Constants for models and defaults
MODEL_LLAMA = "llama3"
MODEL_GEMMA = "gemma"
MODEL_MISTRAL = "mistral"
DEFAULT_MODEL = MODEL_LLAMA
DEFAULT_TEMP = 0.2
DEFAULT_JSON_RETRY = 5
DEFAULT_SYSTEM_PROMPT = (
    "You are a helpful and professional chatbot that assists users with their messages."
)


class ModelResponseError(Exception):
    """Custom exception for invalid model responses."""

    pass


def generate(
    user_prompt: str,
    system_prompt: str,
    model: str = DEFAULT_MODEL,
    temperature: float = DEFAULT_TEMP,
) -> str:
    """
    Generates a response from the specified model.

    Parameters:
        input_text (str): The prompt to send to the model.
        model (str): The model to use for generation.
        temperature (float): Sampling temperature.

    Returns:
        str: The response from the model.

    Raises:
        ModelResponseError: If the response is not a valid mapping.
    """
    resp = ollama.generate(
        model=model,
        prompt=user_prompt,
        system=system_prompt,
        stream=False,
        format="json",
        # keep_alive=10,
        options={
            "temperature": temperature,
        },
    )
    if not isinstance(resp, T.Mapping):
        raise ModelResponseError(
            f"Invalid response from the model for prompt: {user_prompt}"
        )
    out: str = resp.get("response", "")
    if not out:
        raise ModelResponseError(f"Invalid response structure: {resp}")
    return out


def generate_to_dict(
    user_prompt: str,
    system_prompt: str,
    model: str = DEFAULT_MODEL,
    temperature: float = DEFAULT_TEMP,
    retries: int = 3,
) -> T.Dict[str, T.Any]:
    """
    Takes a prompt and returns the response as a dictionary by parsing the JSON.
    Retries parsing up to 'retries' times if JSON decoding fails.

    Parameters:
        input_text (str): The prompt to send to the model.
        model (str): The model to use for generation.
        temperature (float): Sampling temperature.
        retries (int): Number of retry attempts for JSON parsing.

    Returns:
        Dict[str, Any]: The parsed JSON response.

    Raises:
        ModelResponseError: If JSON decoding fails after the specified number of retries.
    """
    # Ensure the prompt instructs the model to respond in JSON
    if "json" not in user_prompt.lower():
        user_prompt += "\nAlways respond with JSON."

    for attempt in range(retries):
        prompt_response = generate(
            user_prompt,
            system_prompt=system_prompt,
            model=model,
            temperature=temperature,
        )
        try:
            return json.loads(prompt_response)
        except json.JSONDecodeError:
            logging.error(
                f"Attempt {attempt + 1}: Could not parse JSON from response: {prompt_response}"
            )
            if attempt < retries - 1:
                user_prompt += f"\nYou must always respond with valid JSON. Your previous response was not valid: {prompt_response}\nRespond to this question with JSON: {user_prompt}"
            else:
                raise ModelResponseError(
                    f"Failed to parse JSON after {retries} attempts. Last response: {prompt_response}"
                )
    raise ModelResponseError("Failed to parse JSON after retries.")


class Conversation:
    def __init__(
        self,
        system_text: str = "You are a helpful chatbot that receives messages and filters or modifies them based on instructions.",
        json_mode: bool = False,
        model: str = DEFAULT_MODEL,
        temperature: float = DEFAULT_TEMP,
    ):
        """
        Initializes a new Conversation instance.

        Parameters:
            system_text (str): The initial system prompt.
            json_mode (bool): Whether to enforce JSON responses.
            model (str): The model to use.
            temperature (float): Sampling temperature.
        """
        self.system_text = system_text
        self.json_mode = json_mode
        self.model = model
        self.ollama_options = ollama.Options(
            temperature=temperature,
        )

        self.messages: T.List[ollama.Message] = []
        self._last_response_full: T.Mapping[str, T.Any] = {}
        self.lock = threading.Lock()  # For thread safety if needed

    def __str__(self):
        return "\n".join([str(msg) for msg in self.messages])

    def __repr__(self):
        return self.__str__()

    @property
    def system_prompt(self) -> ollama.Message:
        """
        Constructs the system prompt message.

        Returns:
            ollama.Message: The system prompt message.
        """
        if self.json_mode and "json" not in self.system_text.lower():
            system_text = (
                f"{self.system_text}\nYou must always respond with valid JSON."
            )
        else:
            system_text = self.system_text
        return ollama.Message(role="system", content=system_text)

    @property
    def last_response(self) -> str:
        """
        Retrieves the content of the last response message.

        Returns:
            str: The content of the last message, or an empty string if no messages exist.
        """
        if not self.messages:
            return ""
        return self.messages[-1].get("content", "")

    @property
    def ollama_chat(self):
        """
        Provides access to the ollama.chat method.
        Defined as a property to enable overloading during testing.

        Returns:
            Callable: The ollama.chat method.
        """
        return ollama.chat

    def chat(self, message_text: str) -> ollama.Message:
        """
        Sends a message to the chat API, updates the message context, and returns the response.

        Parameters:
            message_text (str): The user's message.

        Returns:
            ollama.Message: The model's response message.

        Raises:
            ModelResponseError: If the response structure is invalid.
        """
        logging.info(f"Sending message: {message_text}")
        new_message = ollama.Message(role="user", content=message_text)

        with self.lock:
            # Send the system prompt, previous messages, and the new message to the chat API
            response = self.ollama_chat(
                model=self.model,
                format="json" if self.json_mode else "",
                options=self.ollama_options,
                messages=[self.system_prompt] + self.messages + [new_message],
            )

            if not isinstance(response, T.Mapping):
                logging.error(f"Invalid response type for prompt: {new_message}")
                raise ModelResponseError(
                    f"Invalid response type from the model for prompt: {new_message}"
                )

            resp_dict = response.get("message")
            self._last_response_full = response

            if resp_dict is None:
                logging.error(f"Response missing 'message' key: {response}")
                raise ModelResponseError(
                    f"Invalid response from the model for prompt: {new_message} - {response}"
                )

            required_keys = {"role", "content"}
            if not required_keys.issubset(resp_dict.keys()):
                logging.error(f"Response missing required keys: {resp_dict}")
                raise ModelResponseError(f"Invalid response structure: {resp_dict}")

            # Load the response dict into a Message object
            resp_msg = ollama.Message(**resp_dict)  # type: ignore

            # Append the request and response to the messages context
            self.messages.append(new_message)
            self.messages.append(resp_msg)
            logging.info(f"Received response: {resp_msg.content}")

        return resp_msg
