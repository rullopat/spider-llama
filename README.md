# spider-llama

`spider-llama` is a small Go gateway for routing application LLM requests to private model nodes.

The MVP prioritizes a local `llama.cpp` `llama-server` instance, while keeping the architecture ready for multiple nodes, model aliases, capabilities, tags, and routing policies.

## Project Planning

- [Product requirements](docs/PRD.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Roadmap](docs/ROADMAP.md)
- [Security notes](docs/SECURITY.md)
- [Backlog](docs/backlog/README.md)
- [Architecture decisions](docs/decisions/)

## MVP Shape

```text
client app
  -> spider-llama
  -> router
  -> llama.cpp provider
  -> local llama-server on 127.0.0.1
```

## Start llama.cpp

Example:

```sh
llama-server \
  -m /path/to/model.gguf \
  --host 127.0.0.1 \
  --port 8080 \
  -c 32768 \
  --parallel 1 \
  --jinja
```

## Run spider-llama

```sh
cp configs/spider-llama.example.json spider-llama.json
export SPIDER_LLAMA_TOKEN=change-me
go run ./cmd/spider-llama -config spider-llama.json
```

Build a single binary:

```sh
go build -o bin/spider-llama ./cmd/spider-llama
```

## API

Health is public:

```sh
curl http://127.0.0.1:8088/health
```

Model list requires bearer auth when `SPIDER_LLAMA_TOKEN` is set:

```sh
curl -H "Authorization: Bearer $SPIDER_LLAMA_TOKEN" \
  http://127.0.0.1:8088/v1/models
```

OpenAI-compatible chat request:

```sh
curl -s http://127.0.0.1:8088/v1/chat/completions \
  -H "Authorization: Bearer $SPIDER_LLAMA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "alias:light-text",
    "messages": [
      {"role": "user", "content": "Write one sentence about private LLM routing."}
    ],
    "max_tokens": 64
  }'
```

Router-native request:

```sh
curl -s http://127.0.0.1:8088/v1/route \
  -H "Authorization: Bearer $SPIDER_LLAMA_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "auto",
    "task": "analysis",
    "requirements": {
      "capabilities": ["text", "json"],
      "tags": ["light"]
    }
  }'
```

## Config Concepts

- `nodes`: machines/endpoints that can run models.
- `models`: concrete models on nodes.
- `aliases`: stable names clients can use instead of backend model names.
- `capabilities`: what the model can do, such as `text`, `json`, `tools`, `ocr`, `vision`.
- `tags`: routing hints such as `light`, `reasoning`, `local`, `multilingual`.
- `routes`: policies that map tasks to suitable models.

The first implementation is stateless. Concurrency is enforced per node in memory with `max_concurrency`.
