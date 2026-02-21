# Learning Guide: Customer Support RAG System

A complete walkthrough of every concept and file in this project.

---

## 1. What is RAG (Retrieval-Augmented Generation)?

The core idea: instead of asking an LLM to answer from its training data alone, you **retrieve relevant documents first**, then pass them as context to the LLM alongside the user's question. This gives accurate, grounded answers from your own data.

```
User question → Find relevant docs → Feed docs + question to LLM → Answer
```

**Why RAG?**
- LLMs don't know your private data (FAQs, product docs, help articles)
- Fine-tuning is expensive and slow to update
- RAG lets you update knowledge instantly — just add/edit text files
- Answers are grounded in your actual docs, reducing hallucination

---

## 2. The Data Pipeline (Startup)

When the server starts, it builds a searchable index from your knowledge base. Here's each step:

### Step 1: Load documents (`internal/documents/loader.go`)
- Walks the `knowledge_base/` directory recursively
- Reads all `.txt` and `.md` files, skips empty ones
- Attaches metadata: `filename`, `category`, `source` path

```go
// Key function:
func (l *Loader) LoadFromDirectory(dir string) ([]models.Document, error)
```

Each document becomes a `models.Document` with:
- `ID` — the file path
- `Content` — the full text
- `Metadata` — category (faq, product_docs, help_center), filename, source path

### Step 2: Split into chunks (`internal/documents/splitter.go`)
- Large documents are split into ~500-character chunks (configurable via `CHUNK_SIZE`)
- Chunks **overlap** by ~50 characters so context isn't lost at boundaries

```
Document: "AAAA BBBB CCCC DDDD EEEE"
                          ↓
Chunk 1: "AAAA BBBB CCCC"
Chunk 2: "CCCC DDDD EEEE"   ← "CCCC" overlaps
```

**Why chunk?**
- LLMs have token limits
- Smaller chunks give more precise retrieval (you find the exact relevant paragraph, not a whole 10-page doc)
- Overlap ensures sentences at chunk boundaries aren't lost

### Step 3: Build TF-IDF vocabulary (`internal/embeddings/server.go`)

**TF-IDF** = Term Frequency × Inverse Document Frequency

This is the algorithm that converts text into numbers (vectors) that a computer can compare.

**Term Frequency (TF)**: How often does a word appear in *this* chunk?
```
Chunk: "billing billing support"
TF("billing") = 2/3 = 0.67
TF("support") = 1/3 = 0.33
```

**Inverse Document Frequency (IDF)**: How rare is this word across *all* chunks?
```
If "billing" appears in 5 out of 158 chunks → IDF is high (rare, meaningful)
If "support" appears in 100 out of 158 chunks → IDF is low (common, less meaningful)
```

**TF-IDF = TF × IDF**: Common words like "the" get near-zero weight. Domain-specific words like "billing" get high weight.

**The process:**
1. Tokenize text → split into lowercase words
2. Remove stop words (a, the, is, are, etc.)
3. Count document frequency for each term across all chunks
4. Compute IDF: `log(1 + totalDocs / (1 + docFreq))`
5. For each chunk, compute TF-IDF for every term → produces a vector

**L2 Normalization**: After computing the vector, divide each value by the vector's length. This ensures all vectors have magnitude 1, so cosine similarity works correctly regardless of chunk length.

```go
// The result: each chunk becomes a vector like:
// [0.0, 0.23, 0.0, 0.45, 0.0, 0.12, ...]
//  ↑ each position = one vocabulary term
//  ↑ value = how important that term is to this chunk
```

### Step 4: Store in vector DB (`internal/vectordb/memory.go`)
- All chunks (with their vectors) are stored in a Go slice
- Protected by `sync.RWMutex` for thread-safe concurrent access
- No external database — everything lives in RAM

---

## 3. The Query Pipeline (Per Request)

When a user sends `POST /api/chat`, this pipeline runs:

### Step 1: Embed the query
```go
queryEmbedding := c.embedder.Embed(query)
```
The user's question is converted to a TF-IDF vector using the **same vocabulary** built during startup.

### Step 2: Vector search — Cosine Similarity (`internal/vectordb/search.go`)

Cosine similarity measures the angle between two vectors:

```
                B
               /
              / θ  ← small angle = similar
             /
            A──────→

cos(0°) = 1.0   → identical direction (perfect match)
cos(90°) = 0.0  → perpendicular (unrelated)
```

**Formula:**
```
cosine_similarity = (A · B) / (|A| × |B|)

Where:
  A · B = sum of (a[i] × b[i])     ← dot product
  |A|   = sqrt(sum of a[i]²)       ← magnitude
```

**In Go:**
```go
func CosineSimilarity(a, b []float64) float64 {
    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

The top K (default 5) most similar chunks are returned as sources.

### Step 3: Build prompt (`internal/rag/chain.go`)

The retrieved chunks are formatted into a context block:
```
--- Document 1 (Relevance: 87.34%) ---
Source: knowledge_base/faq/billing.txt
<chunk content here>

--- Document 2 (Relevance: 72.15%) ---
Source: knowledge_base/help_center/troubleshooting.txt
<chunk content here>
```

A system prompt wraps this:
```
You are a helpful customer support assistant. Use the following
pieces of context from our documentation to answer the customer's question.

If you don't find the answer in the context, politely say that you
don't have that information and suggest contacting human support.

Context:
<retrieved chunks>
```

### Step 4: Call LLM (`internal/llm/`)

The context + question + chat history are sent to Claude or OpenAI. The LLM generates a grounded answer based on your docs.

---

## 4. Multi-LLM Provider System

### The Interface Pattern (`internal/llm/client.go`)

```go
type LLMClient interface {
    GenerateResponse(ctx context.Context, query, ragContext string,
        chatHistory []models.ChatMessage) (string, error)
}
```

Both `ClaudeClient` and `OpenAIClient` implement this interface. The RAG chain doesn't know or care which provider it's using.

### Claude Client (`internal/llm/claude.go`)
- Calls `https://api.anthropic.com/v1/messages`
- Auth: `x-api-key` header + `anthropic-version` header
- System prompt is a separate field in the request
- Response: array of `ContentBlock` objects

### OpenAI Client (`internal/llm/openai.go`)
- Calls `https://api.openai.com/v1/chat/completions`
- Auth: `Authorization: Bearer <key>` header
- System prompt is the first message with `role: "system"`
- Response: array of `Choice` objects

### Switching providers
```bash
# In .env:
LLM_PROVIDER=anthropic   # uses Claude
LLM_PROVIDER=openai      # uses GPT
```

In `cmd/server/main.go`, a switch statement picks the right client:
```go
switch cfg.LLMProvider {
case "openai":
    llmClient = llm.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
default:
    llmClient = llm.NewClaudeClient(cfg.AnthropicAPIKey, cfg.ClaudeModel)
}
```

---

## 5. Session Management (`internal/session/manager.go`)

Sessions enable multi-turn conversations (the LLM remembers what you said before).

**How it works:**
1. First request: no `session_id` → server creates one with `crypto/rand` (cryptographically secure random bytes)
2. User and assistant messages are appended to the session
3. Next request: send back the `session_id` → server loads chat history
4. Chat history is passed to the LLM alongside the new question

**TTL cleanup:**
- A background goroutine runs every hour
- Deletes sessions older than `SESSION_TTL_HOURS` (default 24)
- `Stop()` method cleanly shuts down the goroutine via channel close

```go
func (m *Manager) cleanup() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for {
        select {
        case <-m.done:    // Stop() was called
            return
        case <-ticker.C:  // hourly tick
            // delete expired sessions
        }
    }
}
```

---

## 6. API Layer (`internal/api/`)

### HTTP Handlers (`handlers.go`)

**`POST /api/chat`:**
1. Validate method is POST
2. Parse JSON body → `ChatRequest{Query, SessionID}`
3. Get or create session
4. Save user message to session
5. Run RAG pipeline with chat history
6. Save assistant response to session
7. Return JSON → `ChatResponse{SessionID, Answer, Sources}`

**`GET /api/health`:**
Returns `{"status": "healthy", "active_sessions": N, "documents": N}`

### HTTP Server (`server.go`)

**Middleware stack:**
```
Request → Logging → CORS → Handler → Response
```

- **CORS middleware**: Sets `Access-Control-Allow-Origin: *` so any frontend can call the API
- **Logging middleware**: Logs `METHOD /path remote_addr duration` for every request

**Graceful shutdown:**
```go
go func() {
    <-ctx.Done()                    // wait for SIGINT/SIGTERM
    server.Shutdown(shutdownCtx)    // drain existing connections (10s timeout)
}()
```

**Timeouts:**
- Read: 15s (client must send request within 15s)
- Write: 90s (LLM calls can be slow)
- Idle: 60s (keep-alive connection timeout)

---

## 7. Configuration (`internal/config/config.go`)

Custom `.env` parser — no external library needed.

**How it works:**
1. Set all defaults in the `Config` struct
2. Try to open `.env` file (if missing, defaults are used)
3. Scan line by line, skip comments (`#`) and blank lines
4. Split on first `=` → key/value
5. Switch on key, set config field

**All config options:**

| Variable | Default | Purpose |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | Claude API key |
| `SERVER_PORT` | `8080` | HTTP port |
| `KNOWLEDGE_BASE_PATH` | `./knowledge_base` | Docs directory |
| `CHUNK_SIZE` | `500` | Characters per chunk |
| `CHUNK_OVERLAP` | `50` | Overlap between chunks |
| `TOP_K` | `5` | Number of search results |
| `CLAUDE_MODEL` | `claude-sonnet-4-5-20250929` | Anthropic model |
| `SESSION_TTL_HOURS` | `24` | Session expiry |
| `LLM_PROVIDER` | `anthropic` | `anthropic` or `openai` |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `OPENAI_MODEL` | `gpt-4o` | OpenAI model |

---

## 8. Go Patterns & Concepts Used

### Interfaces
```go
type LLMClient interface {
    GenerateResponse(...) (string, error)
}
```
Any struct with a matching `GenerateResponse` method automatically satisfies the interface — no `implements` keyword needed. This is **implicit interface satisfaction**, a core Go concept.

### sync.RWMutex
```go
mu.RLock()   // multiple goroutines can read simultaneously
mu.RUnlock()

mu.Lock()    // only one goroutine can write, blocks all readers
mu.Unlock()
```
Used in `MemoryVectorDB` and `session.Manager` for thread-safe access.

### Context propagation
```go
func (c *ClaudeClient) GenerateResponse(ctx context.Context, ...) (string, error) {
    req, _ := http.NewRequestWithContext(ctx, "POST", url, body)
    // If the HTTP client disconnects, ctx is cancelled,
    // and the LLM API call is aborted
}
```

### Goroutines and channels
```go
done := make(chan struct{})

go func() {    // background goroutine
    select {
    case <-done:   // someone called close(done)
        return
    case <-ticker.C:
        // do periodic work
    }
}()

close(done)  // signals goroutine to stop
```

### Error wrapping
```go
return "", fmt.Errorf("marshal request: %w", err)
//                                       ↑ %w wraps the original error
// Callers can use errors.Is() or errors.As() to inspect the chain
```

### Standard project layout
```
cmd/          → Executables (each subfolder has a main.go)
internal/     → Private packages (Go enforces: can't import from outside this module)
pkg/          → Public packages (can be imported by other projects)
```

---

## 9. Knowledge Base Structure

```
knowledge_base/
├── faq/
│   ├── general.txt        — Platform overview, features, trial info
│   ├── billing.txt        — Plans, pricing, payment, refunds
│   ├── technical.txt      — System requirements, APIs, widget setup
│   └── account.txt        — Signup, password reset, 2FA, roles
├── product_docs/
│   ├── getting_started.txt — Setup wizard, first steps
│   ├── features.txt       — Full feature list
│   ├── api_documentation.txt — REST API reference
│   └── integrations.txt   — Salesforce, Slack, Shopify, etc.
└── help_center/
    ├── troubleshooting.txt — Common issues and fixes
    ├── best_practices.txt  — Tips for ticket management
    └── tutorials.txt       — Step-by-step guides
```

11 documents → ~158 chunks → ~1485 TF-IDF terms

---

## 10. Data Flow Summary

### Startup
```
.env → Config
knowledge_base/*.txt → Loader → []Document
[]Document → Splitter → []Chunk (158 chunks)
[]Chunk.Content → TFIDFEngine.BuildVocabulary() → vocabulary (1485 terms)
[]Chunk.Content → TFIDFEngine.EmbedBatch() → [][]float64 (vectors)
[]Chunk + vectors → MemoryVectorDB.Add()
Config → ClaudeClient or OpenAIClient
All components → RAG Chain → Handlers → HTTP Server → :8080
```

### Per Request
```
POST /api/chat {"query": "How do I reset my password?"}
    ↓
SessionManager.GetOrCreate(session_id)
    ↓
TFIDFEngine.Embed(query) → query vector
    ↓
MemoryVectorDB.Search(query_vector, top_k=5) → 5 relevant chunks
    ↓
Build context string from chunks
    ↓
LLMClient.GenerateResponse(query, context, chat_history)
    ↓
Claude/OpenAI API call → answer
    ↓
Save messages to session
    ↓
{"session_id": "...", "answer": "...", "sources": [...]}
```

---

## 11. Recommended Reading Order

Read the source files in this order to follow the data flow:

| # | File | What you'll learn |
|---|---|---|
| 1 | `pkg/models/types.go` | All data structures used throughout |
| 2 | `internal/config/config.go` | How configuration is loaded from `.env` |
| 3 | `internal/documents/loader.go` | How text files are read from disk |
| 4 | `internal/documents/splitter.go` | How text is chunked with overlap |
| 5 | `internal/embeddings/client.go` | The Embedder interface |
| 6 | `internal/embeddings/server.go` | TF-IDF math (the core algorithm) |
| 7 | `internal/vectordb/search.go` | Cosine similarity formula |
| 8 | `internal/vectordb/memory.go` | How vectors are stored and searched |
| 9 | `internal/llm/client.go` | The LLM provider interface |
| 10 | `internal/llm/claude.go` | How Anthropic API is called |
| 11 | `internal/llm/openai.go` | How OpenAI API is called |
| 12 | `internal/rag/chain.go` | How everything is orchestrated |
| 13 | `internal/session/manager.go` | Conversation state management |
| 14 | `internal/api/handlers.go` | HTTP request handling |
| 15 | `internal/api/server.go` | Server setup and middleware |
| 16 | `cmd/server/main.go` | How it all wires together |
| 17 | `cmd/ingest/main.go` | Dry-run CLI for testing ingestion |

---

## 12. Further Reading

| Topic | What to search |
|---|---|
| RAG | "Retrieval Augmented Generation explained" |
| TF-IDF | "TF-IDF algorithm tutorial with examples" |
| Cosine Similarity | "cosine similarity vectors explained visually" |
| Word Embeddings | "word embeddings vs TF-IDF comparison" |
| Go Interfaces | "Go interface tutorial beginner" |
| Go Concurrency | "Go goroutines channels mutex tutorial" |
| Go Context | "Go context.Context explained" |
| Go Project Layout | "standard Go project layout" |
| Anthropic API | https://docs.anthropic.com/en/api/messages |
| OpenAI API | https://platform.openai.com/docs/api-reference/chat |