package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient implements the Client interface for OpenAI-compatible APIs
type OpenAIClient struct {
	apiKey      string
	baseURL     string
	model       string
	temperature float64
	maxTokens   int
	httpClient  *http.Client
}

// NewOpenAIClient creates a new OpenAI-compatible client
func NewOpenAIClient(apiKey, baseURL, model string, temperature float64, maxTokens int) *OpenAIClient {
	return &OpenAIClient{
		apiKey:      apiKey,
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends a non-streaming chat request
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (string, *RequestStats, error) {
	stats := &RequestStats{
		StartTime: time.Now(),
		Model:     c.model,
	}

	reqBody := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"temperature": c.temperature,
		"max_tokens":  c.maxTokens,
		"stream":      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", stats, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", stats, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", stats, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	stats.HTTPStatus = resp.StatusCode
	stats.EndTime = time.Now()
	stats.Latency = stats.EndTime.Sub(stats.StartTime)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", stats, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", stats, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", stats, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", stats, fmt.Errorf("no choices in response")
	}

	stats.InputTokens = result.Usage.PromptTokens
	stats.OutputTokens = result.Usage.CompletionTokens
	stats.TotalTokens = result.Usage.TotalTokens

	if stats.OutputTokens > 0 && stats.Latency > 0 {
		stats.TokensPerSec = float64(stats.OutputTokens) / stats.Latency.Seconds()
	}

	return result.Choices[0].Message.Content, stats, nil
}

// ChatStream sends a streaming chat request
func (c *OpenAIClient) ChatStream(ctx context.Context, messages []Message) (<-chan StreamChunk, *RequestStats, error) {
	stats := &RequestStats{
		StartTime: time.Now(),
		Model:     c.model,
	}

	reqBody := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"temperature": c.temperature,
		"max_tokens":  c.maxTokens,
		"stream":      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, stats, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, stats, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, stats, fmt.Errorf("failed to send request: %w", err)
	}

	stats.HTTPStatus = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, stats, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	chunks := make(chan StreamChunk, 10)

	go func() {
		defer resp.Body.Close()
		defer close(chunks)

		reader := bufio.NewReader(resp.Body)
		tokenCount := 0
		firstTokenReceived := false

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					chunks <- StreamChunk{Error: fmt.Errorf("stream read error: %w", err)}
				}
				stats.EndTime = time.Now()
				stats.Latency = stats.EndTime.Sub(stats.StartTime)
				stats.OutputTokens = tokenCount

				// Calculate generation time and post-first-token speed
				if firstTokenReceived {
					stats.GenerationTime = stats.EndTime.Sub(stats.FirstTokenTime)
					if tokenCount > 1 && stats.GenerationTime > 0 {
						stats.PostFirstTokenSpeed = float64(tokenCount-1) / stats.GenerationTime.Seconds()
					}
				}

				// Calculate overall tokens per second
				if tokenCount > 0 && stats.Latency > 0 {
					stats.TokensPerSec = float64(tokenCount) / stats.Latency.Seconds()
				}
				chunks <- StreamChunk{Done: true}
				return
			}

			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			// SSE format: "data: {...}"
			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))

			// Check for stream end marker
			if bytes.Equal(data, []byte("[DONE]")) {
				stats.EndTime = time.Now()
				stats.Latency = stats.EndTime.Sub(stats.StartTime)
				stats.OutputTokens = tokenCount

				// Calculate generation time and post-first-token speed
				if firstTokenReceived {
					stats.GenerationTime = stats.EndTime.Sub(stats.FirstTokenTime)
					if tokenCount > 1 && stats.GenerationTime > 0 {
						stats.PostFirstTokenSpeed = float64(tokenCount-1) / stats.GenerationTime.Seconds()
					}
				}

				// Calculate overall tokens per second
				if tokenCount > 0 && stats.Latency > 0 {
					stats.TokensPerSec = float64(tokenCount) / stats.Latency.Seconds()
				}
				chunks <- StreamChunk{Done: true}
				return
			}

			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal(data, &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				content := streamResp.Choices[0].Delta.Content
				tokenCount++ // Approximate token count

				// Track first token timing
				if !firstTokenReceived {
					stats.FirstTokenTime = time.Now()
					stats.TimeToFirstToken = stats.FirstTokenTime.Sub(stats.StartTime)
					firstTokenReceived = true
				}

				chunks <- StreamChunk{Content: content, Done: false}
			}
		}
	}()

	return chunks, stats, nil
}

// GetModel returns the current model
func (c *OpenAIClient) GetModel() string {
	return c.model
}

// SetModel sets the model
func (c *OpenAIClient) SetModel(model string) {
	c.model = model
}

// GetTemperature returns the current temperature
func (c *OpenAIClient) GetTemperature() float64 {
	return c.temperature
}

// SetTemperature sets the temperature
func (c *OpenAIClient) SetTemperature(temp float64) {
	c.temperature = temp
}
