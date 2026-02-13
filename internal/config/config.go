package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AnthropicAPIKey   string
	ServerPort        string
	KnowledgeBasePath string
	ChunkSize         int
	ChunkOverlap      int
	TopK              int
	ClaudeModel       string
	SessionTTLHours   int
	LLMProvider       string
	OpenAIAPIKey      string
	OpenAIModel       string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:        "8080",
		KnowledgeBasePath: "./knowledge_base",
		ChunkSize:         500,
		ChunkOverlap:      50,
		TopK:              5,
		ClaudeModel:       "claude-sonnet-4-5-20250929",
		SessionTTLHours:   24,
		LLMProvider:       "anthropic",
		OpenAIModel:       "gpt-4o",
	}

	file, err := os.Open(".env")
	if err != nil {
		return cfg, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ANTHROPIC_API_KEY":
			cfg.AnthropicAPIKey = value
		case "SERVER_PORT":
			cfg.ServerPort = value
		case "KNOWLEDGE_BASE_PATH":
			cfg.KnowledgeBasePath = value
		case "CHUNK_SIZE":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.ChunkSize = v
			}
		case "CHUNK_OVERLAP":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.ChunkOverlap = v
			}
		case "TOP_K":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.TopK = v
			}
		case "CLAUDE_MODEL":
			cfg.ClaudeModel = value
		case "SESSION_TTL_HOURS":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.SessionTTLHours = v
			}
		case "LLM_PROVIDER":
			cfg.LLMProvider = value
		case "OPENAI_API_KEY":
			cfg.OpenAIAPIKey = value
		case "OPENAI_MODEL":
			cfg.OpenAIModel = value
		}
	}

	return cfg, scanner.Err()
}
