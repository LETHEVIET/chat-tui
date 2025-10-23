package llm

import (
	"context"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// RequestStats tracks statistics for a request
type RequestStats struct {
	StartTime            time.Time
	EndTime              time.Time
	FirstTokenTime       time.Time
	Model                string
	InputTokens          int
	OutputTokens         int
	TotalTokens          int
	TokensPerSec         float64
	TimeToFirstToken     time.Duration
	GenerationTime       time.Duration
	PostFirstTokenSpeed  float64
	Latency              time.Duration
	HTTPStatus           int
	CostEstimate         float64
}

// Client defines the interface for LLM API clients
type Client interface {
	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, messages []Message) (string, *RequestStats, error)

	// ChatStream sends a chat request and streams the response
	ChatStream(ctx context.Context, messages []Message) (<-chan StreamChunk, *RequestStats, error)

	// GetModel returns the current model being used
	GetModel() string

	// SetModel sets the model to use
	SetModel(model string)

	// GetTemperature returns the current temperature
	GetTemperature() float64

	// SetTemperature sets the temperature
	SetTemperature(temp float64)
}
