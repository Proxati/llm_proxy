# builder stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make release

# runtime stage
FROM debian:bookworm-slim
RUN useradd -m -u 10002 proxati
RUN apt-get update \
    && apt-get install -y ca-certificates \
    && update-ca-certificates \
    && apt-get clean all \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/llm_proxy .
RUN chmod +x llm_proxy \
    && ldd llm_proxy \
    && ls -al /app/ \
    && chown -R proxati:nogroup /app
USER proxati
EXPOSE 8080
ENTRYPOINT ["/app/llm_proxy"]
CMD ["run"]