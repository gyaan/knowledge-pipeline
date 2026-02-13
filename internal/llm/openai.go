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

type OpenAIClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *OpenAIClient) GenerateResponse(
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

	messages := make([]models.OpenAIMessage, 0, len(chatHistory)+2)

	messages = append(messages, models.OpenAIMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range chatHistory {
		messages = append(messages, models.OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, models.OpenAIMessage{
		Role:    "user",
		Content: query,
	})

	reqBody := models.OpenAIRequest{
		Model:    c.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	var openAIResp models.OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from OpenAI")
}
