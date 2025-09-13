package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

const (
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
)

type OpenRouterClient struct {
	client *resty.Client
	config *viper.Viper
}

func NewOpenRouterClient(config *viper.Viper) *OpenRouterClient {
	return &OpenRouterClient{
		client: resty.New().
			SetTimeout(30*time.Second).
			SetHeader("Content-Type", "application/json").
			SetHeader("HTTP-Referer", "https://github.com/sstent/fitness-tui").
			SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.GetString("openrouter.apikey"))),
		config: config,
	}
}

type AnalysisRequest struct {
	ActivityData string `json:"activity_data"`
}

type AnalysisResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *OpenRouterClient) AnalyzeActivity(ctx context.Context, prompt string) (string, error) {
	payload := map[string]interface{}{
		"model": c.config.GetString("openrouter.model"),
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	var response AnalysisResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&response).
		Post(openRouterURL)

	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("API error: %s", resp.Status())
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no analysis content in response")
	}

	return response.Choices[0].Message.Content, nil
}
