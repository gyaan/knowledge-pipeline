# Testing Guide

## 1. Build & Static Analysis

```bash
go build ./...
go vet ./...
```

## 2. Dry-Run Ingestion

```bash
go run cmd/ingest/main.go
```

Expected output:
- 11 documents loaded
- ~158 chunks created
- ~1485 unique TF-IDF terms
- 158 non-zero embeddings

## 3. Start the Server

```bash
go run cmd/server/main.go
```

Expected startup logs:
```
Starting Customer Support RAG Server...
Loaded 11 documents
Created 158 chunks
Built TF-IDF vocabulary: 1485 terms
Indexed 158 chunks in vector store
Server starting on port 8080
```

## 4. Health Check

```bash
curl http://localhost:8080/api/health
```

Expected:
```json
{"status":"healthy","active_sessions":0,"documents":158}
```

## 5. Chat Queries

### Basic question (no session)

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "How do I reset my password?"}' | jq .
```

Verify: response has `session_id`, `answer`, and `sources` array.

### Multi-turn conversation (reuse session)

```bash
# First message
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What pricing plans do you offer?"}' | jq .

# Copy the session_id from the response, then:
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Which plan has the best value?", "session_id": "SESSION_ID_HERE"}' | jq .
```

Verify: second response references context from the first turn.

### Sample queries to test different knowledge areas

```bash
# Billing
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Do you offer a free trial?"}' | jq .answer

# Technical
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "How do I install the chat widget on my website?"}' | jq .answer

# Integrations
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "Can I connect Salesforce to CloudSupport Pro?"}' | jq .answer

# Troubleshooting
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "The chat widget is not showing on my site"}' | jq .answer

# Out-of-scope (should gracefully decline)
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the weather today?"}' | jq .answer
```

## 6. Error Cases

### Empty query

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"query": ""}' -w "\n%{http_code}"
```

Expected: `400 Bad Request`

### Wrong HTTP method

```bash
curl -s -X GET http://localhost:8080/api/chat -w "\n%{http_code}"
```

Expected: `405 Method Not Allowed`

### Invalid JSON

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d 'not json' -w "\n%{http_code}"
```

Expected: `400 Bad Request`

## 7. Graceful Shutdown

With the server running, press `Ctrl+C`. Expected logs:

```
Shutting down server...
Shutdown complete
```
