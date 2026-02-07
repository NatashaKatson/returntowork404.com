package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	claudeAPIURL = "https://api.anthropic.com/v1/messages"
	claudeModel  = "claude-sonnet-4-20250514"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type claudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []contentBlock `json:"content"`
	Error   *claudeError   `json:"error,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) GenerateSummary(ctx context.Context, industry, timePeriod string) (string, error) {
	prompt := buildPrompt(industry, timePeriod)

	reqBody := claudeRequest{
		Model:     claudeModel,
		MaxTokens: 2048,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("Claude error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return claudeResp.Content[0].Text, nil
}

func buildPrompt(industry, timePeriod string) string {
	return fmt.Sprintf(`You are a helpful assistant that catches people up on what they missed in their industry while they were away.

A professional in the %s industry has been away for %s and wants to know what major developments, trends, news, and changes they missed.

Please provide a comprehensive but digestible summary covering:
1. **Major News & Events** - Key headlines, mergers, acquisitions, notable company news
2. **Technology & Tool Changes** - New tools, platforms, or technologies that gained adoption
3. **Industry Trends** - Shifting paradigms, emerging practices, changing priorities
4. **Regulatory & Policy Updates** - New laws, regulations, or compliance requirements (if applicable)
5. **Key People & Moves** - Notable leadership changes, influential new voices

Format the response in a friendly, easy-to-read way with clear sections. Use bullet points where appropriate. Keep it informative but not overwhelming - hit the highlights that matter most.

Focus on the most impactful changes that would actually affect someone returning to work in this field.`, industry, timePeriod)
}
