package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ayush/supportiq/internal/ai/parser"
	"github.com/ayush/supportiq/internal/ai/prompt"
	"github.com/ayush/supportiq/internal/ai/provider"
	replyparser "github.com/ayush/supportiq/internal/ai/reply/parser"
	replyprovider "github.com/ayush/supportiq/internal/ai/reply/provider"
	replyprompt "github.com/ayush/supportiq/internal/ai/reply/prompt"
	replyvalidator "github.com/ayush/supportiq/internal/ai/reply/validator"
	aivalidator "github.com/ayush/supportiq/internal/ai/validator"
	"github.com/ayush/supportiq/internal/utils"
)

const apiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

// Client implements provider.Provider and replyprovider.ReplyProvider using the Google Gemini REST API.
type Client struct {
	apiKey           string
	model            string
	maxRetries       int
	maxReplyTokens   int
	replyTemperature float64
	httpClient       *http.Client
}

func NewClient(apiKey, model string, timeout time.Duration, maxRetries int) *Client {
	return &Client{
		apiKey:           apiKey,
		model:            model,
		maxRetries:       maxRetries,
		maxReplyTokens:   1024,
		replyTemperature: 0.3,
		httpClient:       &http.Client{Timeout: timeout},
	}
}

// NewClientWithReplyConfig creates a Client with explicit reply generation parameters.
func NewClientWithReplyConfig(apiKey, model string, timeout time.Duration, maxRetries, maxReplyTokens int, replyTemperature float64) *Client {
	c := NewClient(apiKey, model, timeout, maxRetries)
	c.maxReplyTokens = maxReplyTokens
	c.replyTemperature = replyTemperature
	return c
}

// ─── Internal request / response types ───────────────────────────────────────

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// ─── provider.Provider implementation ────────────────────────────────────────

// Analyze calls the Gemini API with automatic retries and structured logging.
func (c *Client) Analyze(ctx context.Context, req provider.AnalysisRequest) (*provider.AnalysisResult, error) {
	promptText := prompt.BuildTicketAnalysisPrompt(
		req.Subject, req.Description, req.CustomerName, req.Category, req.Priority,
	)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			utils.Logger.WithField("attempt", attempt).
				WithField("model", c.model).
				Info("AI: Retrying Gemini API call")
		}

		result, err := c.callOnce(ctx, promptText)
		if err == nil {
			return result, nil
		}
		lastErr = err
		utils.Logger.WithError(err).
			WithField("attempt", attempt).
			Warn("AI: Gemini API call failed")
	}

	return nil, fmt.Errorf("all %d attempt(s) failed, last error: %w", c.maxRetries+1, lastErr)
}

// ─── replyprovider.ReplyProvider implementation ───────────────────────────────

// GenerateReply calls the Gemini API to generate a customer support reply using RAG context.
func (c *Client) GenerateReply(ctx context.Context, req replyprovider.ReplyRequest) (*replyprovider.ReplyResult, error) {
	promptText := replyprompt.BuildReplyPrompt(req)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			utils.Logger.WithField("attempt", attempt).
				WithField("model", c.model).
				Info("AI Reply: Retrying Gemini API call")
		}

		rawText, err := c.callAPI(ctx, promptText, c.replyTemperature, c.maxReplyTokens)
		if err != nil {
			lastErr = err
			utils.Logger.WithError(err).WithField("attempt", attempt).Warn("AI Reply: Gemini API call failed")
			continue
		}

		parsed, err := replyparser.Parse(rawText)
		if err != nil {
			lastErr = err
			utils.Logger.WithError(err).Warn("AI Reply: Response parse failed")
			continue
		}

		if err := replyvalidator.Validate(parsed); err != nil {
			lastErr = err
			utils.Logger.WithError(err).Warn("AI Reply: Response validation failed")
			continue
		}

		return &replyprovider.ReplyResult{
			Reply:      parsed.Reply,
			Confidence: parsed.Confidence,
		}, nil
	}

	return nil, fmt.Errorf("all %d attempt(s) failed, last error: %w", c.maxRetries+1, lastErr)
}

// ─── Shared HTTP layer ────────────────────────────────────────────────────────

// callOnce wraps callAPI with analysis-specific parsing and validation.
func (c *Client) callOnce(ctx context.Context, promptText string) (*provider.AnalysisResult, error) {
	rawText, err := c.callAPI(ctx, promptText, 0.1, 512)
	if err != nil {
		return nil, err
	}

	parsed, err := parser.Parse(rawText)
	if err != nil {
		utils.Logger.WithError(err).Warn("AI: Response parse failed")
		return nil, err
	}

	if err := aivalidator.Validate(parsed); err != nil {
		utils.Logger.WithError(err).Warn("AI: Response validation failed")
		return nil, err
	}

	return &provider.AnalysisResult{
		Category:        parsed.Category,
		Priority:        parsed.Priority,
		Sentiment:       parsed.Sentiment,
		RecommendedTeam: parsed.RecommendedTeam,
		Confidence:      parsed.Confidence,
		Summary:         parsed.Summary,
		Tags:            parsed.Tags,
	}, nil
}

// callAPI performs a single HTTP call to the Gemini API and returns the raw text response.
// All callers (analysis, reply generation) share this implementation.
func (c *Client) callAPI(ctx context.Context, promptText string, temperature float64, maxOutputTokens int) (string, error) {
	start := time.Now()

	reqBody := geminiRequest{
		Contents:         []geminiContent{{Parts: []geminiPart{{Text: promptText}}}},
		GenerationConfig: geminiGenerationConfig{Temperature: temperature, MaxOutputTokens: maxOutputTokens},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build URL — apiKey is intentionally kept out of log output
	url := fmt.Sprintf(apiBaseURL+"?key=%s", c.model, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to build HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		utils.Logger.WithField("status", resp.StatusCode).
			WithField("latency_ms", latency.Milliseconds()).
			Warn("AI: Gemini API returned non-200")
		return "", fmt.Errorf("gemini API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode Gemini response body: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned an empty candidate list")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text

	utils.Logger.WithField("latency_ms", latency.Milliseconds()).
		WithField("model", c.model).
		WithField("tokens_total", geminiResp.UsageMetadata.TotalTokenCount).
		WithField("tokens_prompt", geminiResp.UsageMetadata.PromptTokenCount).
		WithField("tokens_response", geminiResp.UsageMetadata.CandidatesTokenCount).
		Info("AI: Gemini response received")

	return rawText, nil
}

