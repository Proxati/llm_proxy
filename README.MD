# LLM Proxy
A high-performance, easy-to-install proxy server designed to intercept and modify requests to LLM
APIs like OpenAI. With a single compiled binary and no external runtime dependencies, it’s easy to
deploy and operate.


## Features

* Easy Installation: Easy to deploy and run with a single compiled binary.
* High Performance: Written in Go, the proxy is fast and efficient.
* Caching and Logging: Save all requests and responses for various use cases:
  * Fine-tune Models
  * Security Auditing
  * Debugging
  * Academic Reproducibility

## How to install and run the proxy

1. Install Go
2. Run `go install github.com/proxati/llm_proxy@latest`
3. The binary will be stored in your `$GOPATH/bin` directory.
4. Verify installation with `llm_proxy --help`.
5. Set your `HTTP_PROXY` and `HTTPS_PROXY` environment variables to `http://localhost:8080`.
6. Use the OpenAI API as you normally would.

### Running the proxy server
```bash
$ llm_proxy run --verbose
```

### Using cURL to query, and use the proxy
(Set your OpenAI API key in the header)
```bash
$ curl \
    -x http://localhost:8080 \
    -X GET \
    -H "Authorization: Bearer sk-XXXXXXX" \
    http://api.openai.com/v1/models
```
Note: This example uses `http://api.openai.com/...` instead of https:// because the proxy handles
SSL termination and upgrades the outbound request to `https://`. See the TLS section for more
details.

### Using the proxy with the OpenAI Python client

In this example, we are changing the `base_url` to connect via `http` so the proxy can MITM the
connection without needing to add a self-signed cert to the Python client.

```python
import httpx
from openai import OpenAI

proxies = {
    "http://": "http://localhost:8080",
    "https://": "http://localhost:8080",
}

client = OpenAI(
    # max_retries=0,
    base_url="http://api.openai.com/v1",
    http_client=httpx.Client(
        proxies=proxies,
    ),
)

chat_completion = client.chat.completions.create(
    messages=[
        {
            "role": "user",
            "content": "Hello, you are amazing.",
        }
    ],
    model="gpt-3.5-turbo",
)
```
More info here: [httpx proxy config](https://www.python-httpx.org/advanced/#client-instances)

## TLS / HTTPs Support

Requests sent to `http://api.openai.com` are upgraded to `https://api.openai.com` by the proxy. If
you need to send requests directly to `https://api.openai.com`, you must add the self-signed cert
to your trust store or disable TLS validation (not recommended).

The proxy generates a certificate at `~/.mitmproxy/mitmproxy-ca-cert.pem`. To use a different
directory, use the `--ca_dir` flag when starting the proxy daemon.

More info on self-signed certs and MITM:
[https://docs.mitmproxy.org/stable/concepts-certificates/]


### Generating and Trusting a Self-Signed Cert on MacOS

```bash
# Create a directory for the cert files
$ mkdir -p ~/.mitmproxy
$ cd ~/.mitmproxy

# you only need to generate this cert if you do not want the llm_proxy to generate it for you
$ openssl genrsa -out mitmproxy-ca-cert.key 2048
# this self-signed cert expires in 10 years, and I hope you are using something else by that point
$ openssl req -x509 -new -nodes -key mitmproxy-ca-cert.key -sha256 -days 3650 -out mitmproxy-ca-cert.pem

# Trust the CA
$ sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain mitmproxy-ca-cert.pem
```

### Using cURL with the Proxy and Self-Signed Cert
(Set your OpenAI API key in the header)
```bash
$ curl \
    -x http://localhost:8080 \
    --cacert ~/.mitmproxy/mitmproxy-ca.pem \
    -X GET \
    -H "Authorization: Bearer sk-XXXXXXX" \
    https://api.openai.com/v1/models
```

