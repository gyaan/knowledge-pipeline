package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/gyaan/knowledge-pipeline/internal/embeddings"
	"github.com/gyaan/knowledge-pipeline/internal/llm"
	"github.com/gyaan/knowledge-pipeline/internal/vectordb"
	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type Chain struct {
	embedder  embeddings.Embedder
	vectorDB  *vectordb.MemoryVectorDB
	llmClient llm.LLMClient
	topK      int
}

func NewChain(
	embedder embeddings.Embedder,
	vectorDB *vectordb.MemoryVectorDB,
	llmClient llm.LLMClient,
	topK int,
) *Chain {
	return &Chain{
		embedder:  embedder,
		vectorDB:  vectorDB,
		llmClient: llmClient,
		topK:      topK,
	}
}

// Query processes a user query through the RAG pipeline.
func (c *Chain) Query(
	ctx context.Context,
	query string,
	chatHistory []models.ChatMessage,
) (*models.ChatResponse, error) {

	// 1. Generate embedding for query
	queryEmbedding := c.embedder.Embed(query)

	// 2. Search vector database
	sources := c.vectorDB.Search(queryEmbedding, c.topK)

	// 3. Build context from sources
	var contextBuilder strings.Builder
	for i, source := range sources {
		contextBuilder.WriteString(fmt.Sprintf("\n\n--- Document %d (Relevance: %.2f%%) ---\n",
			i+1, source.Score*100))
		contextBuilder.WriteString(fmt.Sprintf("Source: %s\n", source.Metadata["source"]))
		contextBuilder.WriteString(source.Content)
	}

	// 4. Generate response with LLM
	answer, err := c.llmClient.GenerateResponse(
		ctx,
		query,
		contextBuilder.String(),
		chatHistory,
	)
	if err != nil {
		return nil, fmt.Errorf("generate response: %w", err)
	}

	return &models.ChatResponse{
		Answer:  answer,
		Sources: sources,
	}, nil
}
