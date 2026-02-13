package llm

import (
	"context"

	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

// LLMClient defines the interface for language model providers.
type LLMClient interface {
	GenerateResponse(ctx context.Context, query, ragContext string, chatHistory []models.ChatMessage) (string, error)
}
