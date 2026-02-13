package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type ClaudeClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

func NewClaudeClient(apiKey, model string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateResponse generates a response using the Claude API.
func (c *ClaudeClient) GenerateResponse(
	ctx context.Context,
	query string,
	ragContext string,
	chatHistory []models.ChatMessage,
) (string, error) {

	systemPrompt := fmt.Sprintf(`You are a helpful customer support assistant. Use the following pieces of context from our documentation to answer the customer's question.

If you don't find the answer in the context, politely say that you don't have that information and suggest contacting human support.

Always be professional, friendly, and concise. If relevant, cite which document or section you're referencing.

Context:
%s`, ragContext)

	messages := make([]models.ClaudeMessage, 0, len(chatHistory)+1)

	for _, msg := range chatHistory {
		messages = append(messages, models.ClaudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, models.ClaudeMessage{
		Role:    "user",
		Content: query,
	})

	reqBody := models.ClaudeRequest{
		Model:     c.model,
		MaxTokens: 1024,
		Messages:  messages,
		System:    systemPrompt,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp models.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(claudeResp.Content) > 0 {
		return claudeResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("empty response from Claude")
}
