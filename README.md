# Customer Support RAG

[![CI](https://github.com/gyaan/knowledge-pipeline/actions/workflows/ci.yml/badge.svg)](https://github.com/gyaan/knowledge-pipeline/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Zero dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen)](go.mod)

A production-ready Retrieval-Augmented Generation (RAG) system built in pure Go. It serves API endpoints for a frontend to answer customer queries using a knowledge base of FAQs, product docs, and help center articles as context.

## Features

- Pure Go TF-IDF embeddings (no Python, no external embedding service)
- In-memory vector store with cosine similarity search
- Multi-LLM support: Anthropic Claude and OpenAI GPT
- Multi-turn conversation sessions with configurable TTL
- Zero external dependencies — pure `net/http` + standard library only
- Graceful shutdown with signal handling

## Quick Start

### 1. Clone and configure

```bash
git clone https://github.com/gyaan/knowledge-pipeline.git
cd knowledge-pipeline
cp .env_example .env
```

Edit `.env` and set your API key:

```
ANTHROPIC_API_KEY=your_key_here
```

Or to use OpenAI instead:

```
LLM_PROVIDER=openai
OPENAI_API_KEY=your_key_here
```

### 2. Run the server

```bash
go run cmd/server/main.go
```

### 3. Try it out

```bash
# Health check
curl http://localhost:8080/api/health

# Ask a question
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "How do I reset my password?"}' | jq .
```

## API Endpoints

### `POST /api/chat`

```json
// Request
{"query": "How do I reset my password?", "session_id": "optional-for-multi-turn"}

// Response
{"session_id": "...", "answer": "...", "sources": [...]}
```

### `GET /api/health`

```json
{"status": "healthy", "active_sessions": 0, "documents": 158}
```

## Architecture

```
Startup: Load docs → Split → Build TF-IDF vocab → Embed chunks → Store in vector DB
Request: POST /api/chat → Embed query → Vector search → Build context → LLM → Response
```

### Project Structure

```
cmd/
  ingest/       — Dry-run CLI (loads KB, chunks, builds TF-IDF, prints stats)
  server/       — HTTP API server with graceful shutdown
internal/
  api/          — HTTP router and handlers
  config/       — .env configuration loader
  documents/    — Document loader and text splitter
  embeddings/   — TF-IDF embedding engine
  llm/          — LLM client interface + Anthropic/OpenAI implementations
  rag/          — RAG chain orchestrator
  session/      — Conversation session manager
  vectordb/     — In-memory vector store with cosine similarity
pkg/
  models/       — Shared types (Document, ChatRequest/Response, API types)
knowledge_base/
  faq/          — General, billing, technical, account FAQs
  product_docs/ — Getting started, features, API docs, integrations
  help_center/  — Troubleshooting, best practices, tutorials
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | Claude API key (required when provider is `anthropic`) |
| `SERVER_PORT` | `8080` | HTTP server port |
| `KNOWLEDGE_BASE_PATH` | `./knowledge_base` | Path to knowledge base directory |
| `CHUNK_SIZE` | `500` | Text chunk size in characters |
| `CHUNK_OVERLAP` | `50` | Overlap between chunks |
| `TOP_K` | `5` | Number of results for vector search |
| `CLAUDE_MODEL` | `claude-sonnet-4-5-20250929` | Anthropic model name |
| `SESSION_TTL_HOURS` | `24` | Session expiry time |
| `LLM_PROVIDER` | `anthropic` | LLM provider: `anthropic` or `openai` |
| `OPENAI_API_KEY` | — | OpenAI API key (required when provider is `openai`) |
| `OPENAI_MODEL` | `gpt-4o` | OpenAI model name |

## Development

```bash
go build ./...                  # Build all packages
go test ./...                   # Run all tests
go vet ./...                    # Static analysis
go run cmd/ingest/main.go       # Dry-run ingestion (no LLM needed)
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bug reports and feature requests go through the [issue tracker](https://github.com/gyaan/knowledge-pipeline/issues/new/choose).

## Security

See [SECURITY.md](SECURITY.md) for the vulnerability disclosure policy and security considerations for self-hosted deployments.

## License

MIT
