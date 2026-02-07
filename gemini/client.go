package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	geminiModel  = "gemini-2.0-flash"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type geminiRequest struct {
	Contents         []content         `json:"contents"`
	GenerationConfig *generationConfig `json:"generationConfig,omitempty"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []candidate   `json:"candidates"`
	Error      *geminiError  `json:"error,omitempty"`
}

type candidate struct {
	Content content `json:"content"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
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
	log.Printf("[Gemini] Generating summary: industry=%s, period=%s", industry, timePeriod)
	prompt := buildPrompt(industry, timePeriod)

	reqBody := geminiRequest{
		Contents: []content{
			{
				Parts: []part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &generationConfig{
			MaxOutputTokens: 2048,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Gemini uses API key as query parameter
	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	log.Printf("[Gemini] Sending request to Gemini API (model: %s)", geminiModel)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Gemini] HTTP request failed: %v", err)
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Gemini] API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		log.Printf("[Gemini] API error: code=%d, message=%s, status=%s",
			geminiResp.Error.Code, geminiResp.Error.Message, geminiResp.Error.Status)
		return "", fmt.Errorf("Gemini error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 {
		log.Printf("[Gemini] Empty candidates in response")
		return "", fmt.Errorf("empty response from Gemini")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in Gemini response")
	}

	summary := geminiResp.Candidates[0].Content.Parts[0].Text
	log.Printf("[Gemini] Successfully generated summary (%d characters)", len(summary))
	return summary, nil
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
