package models

import "time"

type Document struct {
	ID        string
	Content   string
	Metadata  map[string]string
	Embedding []float64
}

type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatRequest struct {
	SessionID string `json:"session_id,omitempty"`
	Query     string `json:"query"`
}

type ChatResponse struct {
	SessionID string           `json:"session_id"`
	Answer    string           `json:"answer"`
	Sources   []SourceDocument `json:"sources"`
}

type SourceDocument struct {
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Score    float64           `json:"score"`
}

type Session struct {
	ID        string
	Messages  []ChatMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// OpenAI API types

type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string         `json:"id"`
	Choices []OpenAIChoice `json:"choices"`
}

type OpenAIChoice struct {
	Index   int           `json:"index"`
	Message OpenAIMessage `json:"message"`
}
