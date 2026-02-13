package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/gyaan/knowledge-pipeline/internal/api"
	"github.com/gyaan/knowledge-pipeline/internal/config"
	"github.com/gyaan/knowledge-pipeline/internal/documents"
	"github.com/gyaan/knowledge-pipeline/internal/embeddings"
	"github.com/gyaan/knowledge-pipeline/internal/llm"
	"github.com/gyaan/knowledge-pipeline/internal/rag"
	"github.com/gyaan/knowledge-pipeline/internal/session"
	"github.com/gyaan/knowledge-pipeline/internal/vectordb"
)

func main() {
	log.Println("Starting Customer Support RAG Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate API key for selected provider
	var llmClient llm.LLMClient
	switch cfg.LLMProvider {
	case "openai":
		if cfg.OpenAIAPIKey == "" {
			log.Fatal("OPENAI_API_KEY must be set in .env when LLM_PROVIDER=openai")
		}
		llmClient = llm.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
		log.Printf("Using OpenAI provider (model: %s)", cfg.OpenAIModel)
	default:
		if cfg.AnthropicAPIKey == "" || cfg.AnthropicAPIKey == "your_anthropic_api_key_here" {
			log.Fatal("ANTHROPIC_API_KEY must be set in .env when LLM_PROVIDER=anthropic")
		}
		llmClient = llm.NewClaudeClient(cfg.AnthropicAPIKey, cfg.ClaudeModel)
		log.Printf("Using Anthropic provider (model: %s)", cfg.ClaudeModel)
	}

	// Load documents
	log.Printf("Loading knowledge base from %s...", cfg.KnowledgeBasePath)
	loader := documents.NewLoader()
	docs, err := loader.LoadFromDirectory(cfg.KnowledgeBasePath)
	if err != nil {
		log.Fatalf("Failed to load documents: %v", err)
	}
	log.Printf("Loaded %d documents", len(docs))

	// Split into chunks
	splitter := documents.NewSplitter(cfg.ChunkSize, cfg.ChunkOverlap)
	chunks := splitter.Split(docs)
	log.Printf("Created %d chunks", len(chunks))

	// Build TF-IDF vocabulary and embed chunks
	tfidf := embeddings.NewTFIDFEngine()
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}
	tfidf.BuildVocabulary(texts)
	log.Printf("Built TF-IDF vocabulary: %d terms", tfidf.VocabSize())

	vecs := tfidf.EmbedBatch(texts)
	for i := range chunks {
		chunks[i].Embedding = vecs[i]
	}

	// Store in vector database
	vectorDB := vectordb.NewMemoryVectorDB()
	vectorDB.Add(chunks)
	log.Printf("Indexed %d chunks in vector store", vectorDB.Count())

	// Initialize remaining components
	ragChain := rag.NewChain(tfidf, vectorDB, llmClient, cfg.TopK)
	sessionManager := session.NewManager(cfg.SessionTTLHours)
	handlers := api.NewHandlers(ragChain, sessionManager, vectorDB)
	server := api.NewServer(handlers, cfg.ServerPort)

	// Graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Start(ctx); err != nil {
		log.Printf("Server stopped: %v", err)
	}

	sessionManager.Stop()
	log.Println("Shutdown complete")
}
