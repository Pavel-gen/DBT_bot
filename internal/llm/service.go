package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Service struct {
	apiKey       string
	apiURL       string
	httpClient   *http.Client
	promptLoader *PromptLoader
}

type Message struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

type LLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewService(promptLoader *PromptLoader) *Service {
	return &Service{
		apiKey:       os.Getenv("OPENROUTER_API_KEY"),
		apiURL:       "https://openrouter.ai/api/v1/chat/completions",
		promptLoader: promptLoader,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Чуть больше, чем в JS (1200 токенов)
		},
	}
}

// Call отправляет запрос к OpenRouter и возвращает сырой ответ
func (s *Service) Call(ctx context.Context, systemKey, userText string, maxTokens int, temperature float64) (string, error) {
	systemPrompt := s.promptLoader.Get(systemKey)
	if systemPrompt == "" {
		return "", fmt.Errorf("prompt '%s' not found", systemKey)
	}

	reqBody := LLMRequest{
		Model: "deepseek/deepseek-v3.2",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userText},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	// === РЕТРАИ: 3 попытки с экспоненциальной задержкой ===
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		// Создаём контекст с таймаутом 90 секунд (больше, чем раньше)
		callCtx, cancel := context.WithTimeout(ctx, 90*time.Second)

		resp, err := s.doHTTPRequest(callCtx, jsonData)
		cancel() // всегда освобождаем ресурсы

		if err == nil {
			return resp, nil // успех
		}

		lastErr = err
		log.Printf("⚠️ [LLM] Attempt %d failed: %v", attempt, err)

		if attempt < 3 {
			// Ждём: 2с → 4с → (след. попытка)
			backoff := time.Duration(1<<attempt) * time.Second
			select {
			case <-time.After(backoff):
				// продолжаем цикл
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("all %d attempts failed: %w", 3, lastErr)
}

func (s *Service) doHTTPRequest(ctx context.Context, jsonData []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://your-bot-url.com")
	req.Header.Set("X-Title", "MyGoBot")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(llmResp.Choices) == 0 || llmResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from LLM")
	}

	return llmResp.Choices[0].Message.Content, nil
}
