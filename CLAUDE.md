# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**customer-support-rag** is a production-ready RAG (Retrieval-Augmented Generation) system that serves API endpoints for a frontend to answer customer queries. It uses a knowledge base of FAQs, product docs, and help center articles as context, with Claude as the LLM and a pure Go TF-IDF engine for embeddings with an in-memory vector store for semantic search.

- **Repository**: https://github.com/gyaan/knowledge-pipeline.git
- **License**: MIT
- **Language**: Go 1.25.6 (macOS ARM64)
- **External dependencies**: None (pure standard library)

## Build & Run Commands

```bash
go build ./...                                      # build all packages
go test ./...                                       # run all tests
go test -run TestName ./path/to/package             # run a single test
go vet ./...                                        # static analysis

# Dry-run ingestion (loads KB, chunks, builds TF-IDF, prints stats)
go run cmd/ingest/main.go

# Run the API server (ingests KB on startup, serves on :8080)
go run cmd/server/main.go
```

## Architecture

```
Startup: Load docs → Split → Build TF-IDF vocab → Embed chunks → Store in vector DB
Request: POST /api/chat → Embed query → Vector search → Build context → Claude LLM → Response
```

### Entry Points
- **cmd/ingest/** — Dry-run CLI: loads knowledge base, chunks, builds TF-IDF vocab, prints stats (no server)
- **cmd/server/** — Ingests KB at startup, then serves HTTP API with graceful shutdown (SIGINT/SIGTERM)

### Internal Packages (data flow order)
1. **config/** — Loads .env: API key, server port, KB path, chunk settings, model, session TTL
2. **documents/** — `loader.go` reads .txt files with category/filename metadata, skips empty; `splitter.go` chunks with overlap
3. **embeddings/** — `client.go` defines `Embedder` interface; `server.go` implements pure Go TF-IDF engine
4. **vectordb/** — `memory.go` in-memory store (no disk persistence); `search.go` cosine similarity
5. **llm/** — `client.go` defines `LLMClient` interface; `claude.go` Anthropic client; `openai.go` OpenAI client
6. **rag/** — `chain.go` orchestrates: query → embed → vector search → build prompt → call LLM → return
7. **session/** — `manager.go` manages conversation history with configurable TTL, crypto/rand IDs, stoppable cleanup
8. **api/** — `server.go` HTTP router with graceful shutdown; `handlers.go` defines ChatHandler + HealthHandler

### Shared Types
- **pkg/models/** — `types.go` defines shared structs (Document, ChatRequest/Response, Session, Claude/OpenAI API types)

### API Endpoints
- `POST /api/chat` — `{"query": "...", "session_id": "..."}` → `{"session_id": "...", "answer": "...", "sources": [...]}`
- `GET /api/health` — `{"status": "healthy", "active_sessions": N, "documents": N}`

### Knowledge Base (`knowledge_base/`)
- **faq/** — general, billing, technical, account FAQs
- **product_docs/** — getting started, features, API docs, integrations
- **help_center/** — troubleshooting, best practices, tutorials

## Environment Variables (.env)

```
ANTHROPIC_API_KEY=         # Claude API key (required when LLM_PROVIDER=anthropic)
SERVER_PORT=8080           # HTTP server port
KNOWLEDGE_BASE_PATH=./knowledge_base
CHUNK_SIZE=500
CHUNK_OVERLAP=50
TOP_K=5
CLAUDE_MODEL=claude-sonnet-4-5-20250929
SESSION_TTL_HOURS=24
LLM_PROVIDER=anthropic     # anthropic or openai
OPENAI_API_KEY=            # OpenAI API key (required when LLM_PROVIDER=openai)
OPENAI_MODEL=gpt-4o        # OpenAI model name
```

## Conventions
- Zero external dependencies — pure `net/http` + standard library only
- In-memory vector store (no external DB, no disk persistence)
- Pure Go TF-IDF embeddings (no Python, no external embedding service)
- Standard Go project layout: `cmd/`, `internal/`, `pkg/`
- Configuration via `.env` file (gitignored)
- Graceful shutdown with signal handling
