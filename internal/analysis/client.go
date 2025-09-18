package analysis

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sstent/fitness-tui/internal/circuitbreaker"

	"github.com/go-resty/resty/v2"
	"github.com/sstent/fitness-tui/internal/config"
)

// OpenRouterClient handles communication with the OpenRouter API
// It manages request retries, circuit breaking, and authentication
type OpenRouterClient struct {
	client         *resty.Client                  // HTTP client instance
	config         *config.Config                 // Application configuration
	circuitBreaker *circuitbreaker.CircuitBreaker // Circuit breaker for API availability
}

func NewOpenRouterClient(cfg *config.Config) *OpenRouterClient {
	timeout := cfg.OpenRouter.Timeout
	if timeout == 0 {
		// Fallback to 30s if timeout is not set
		timeout = 30 * time.Second
	}

	cb := circuitbreaker.New(5, 30*time.Second)
	return &OpenRouterClient{
		client: resty.New().
			SetBaseURL(cfg.OpenRouter.BaseURL).
			SetTimeout(timeout).
			SetHeader("Content-Type", "application/json").
			SetHeader("HTTP-Referer", "https://github.com/sstent/fitness-tui").
			SetHeader("Authorization", fmt.Sprintf("Bearer %s", cfg.OpenRouter.APIKey)),
		config:         cfg,
		circuitBreaker: cb,
	}
}

type AnalysisResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *OpenRouterClient) AnalyzeActivity(ctx context.Context, params PromptParams) (string, error) {
	// Generate prompt using the prompts package
	var err error
	prompt := GeneratePrompt(params)
	payload := map[string]interface{}{
		"model": c.config.OpenRouter.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	var response AnalysisResponse
	var resp *resty.Response

	// Check circuit breaker state
	if !c.circuitBreaker.AllowRequest() {
		return "", fmt.Errorf("API unavailable (circuit breaker open)")
	}

	// Retry with exponential backoff and jitter
	maxRetries := 5
	baseDelay := 500 * time.Millisecond
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		// Check if we should abort due to circuit breaker
		if !c.circuitBreaker.AllowRequest() {
			return "", fmt.Errorf("API unavailable (circuit breaker open)")
		}
		resp, err = c.client.R().
			SetContext(ctx).
			SetBody(payload).
			SetResult(&response).
			Post("/chat/completions")

		if err == nil && resp.IsSuccess() {
			break
		}

		// Handle rate limiting (429)
		if resp != nil && resp.StatusCode() == 429 {
			retryAfter := resp.Header().Get("Retry-After")
			if retryAfter != "" {
				if delay, err := time.ParseDuration(retryAfter + "s"); err == nil {
					time.Sleep(delay)
					continue
				}
			}
		}

		// If context cancelled, break immediately
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// Calculate next backoff with jitter
		delay := baseDelay * time.Duration(1<<uint(i))         // Exponential: 500ms, 1s, 2s, 4s, 8s
		jitter := time.Duration(rand.Int63n(int64(delay / 2))) // Up to 50% jitter
		totalDelay := delay + jitter

		// Store last error for final return
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", i+1, err)
		} else {
			lastErr = fmt.Errorf("attempt %d: API error %s", i+1, resp.Status())
		}

		time.Sleep(totalDelay)
	}

	if err != nil || resp.IsError() {
		c.circuitBreaker.RecordFailure()
		return "", fmt.Errorf("API request failed: %w", lastErr)
	}

	c.circuitBreaker.RecordSuccess()

	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty analysis content in API response")
	}

	return response.Choices[0].Message.Content, nil
}
