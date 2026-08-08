package groq

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
	replyprompt "github.com/ayush/supportiq/internal/ai/reply/prompt"
	replyprovider "github.com/ayush/supportiq/internal/ai/reply/provider"
	replyvalidator "github.com/ayush/supportiq/internal/ai/reply/validator"
	aivalidator "github.com/ayush/supportiq/internal/ai/validator"
	"github.com/ayush/supportiq/internal/utils"
)

const apiBaseURL = "https://api.groq.com/openai/v1/chat/completions"

// Client implements provider.Provider and replyprovider.ReplyProvider using Groq's API.
// Groq is free, fast, and OpenAI-compatible.
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

func NewClientWithReplyConfig(apiKey, model string, timeout time.Duration, maxRetries, maxReplyTokens int, replyTemperature float64) *Client {
	c := NewClient(apiKey, model, timeout, maxRetries)
	c.maxReplyTokens = maxReplyTokens
	c.replyTemperature = replyTemperature
	return c
}

type groqRequest struct {
	Model       string        `json:"model"`
	Messages    []groqMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
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
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Analyze(ctx context.Context, req provider.AnalysisRequest) (*provider.AnalysisResult, error) {
	promptText := prompt.BuildTicketAnalysisPrompt(
		req.Subject, req.Description, req.CustomerName, req.Category, req.Priority,
	)
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			utils.Logger.WithField("attempt", attempt).Info("AI: Retrying Groq API call")
		}
		result, err := c.callOnce(ctx, promptText)
		if err == nil {
			return result, nil
		}
		lastErr = err
		utils.Logger.WithError(err).WithField("attempt", attempt).Warn("AI: Groq API call failed")
	}
	return nil, fmt.Errorf("all %d attempt(s) failed: %w", c.maxRetries+1, lastErr)
}

func (c *Client) GenerateReply(ctx context.Context, req replyprovider.ReplyRequest) (*replyprovider.ReplyResult, error) {
	promptText := replyprompt.BuildReplyPrompt(req)
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		rawText, err := c.callAPI(ctx, promptText, c.replyTemperature, c.maxReplyTokens)
		if err != nil {
			lastErr = err
			continue
		}
		parsed, err := replyparser.Parse(rawText)
		if err != nil {
			lastErr = err
			continue
		}
		if err := replyvalidator.Validate(parsed); err != nil {
			lastErr = err
			continue
		}
		return &replyprovider.ReplyResult{Reply: parsed.Reply, Confidence: parsed.Confidence}, nil
	}
	return nil, fmt.Errorf("all %d attempt(s) failed: %w", c.maxRetries+1, lastErr)
}

func (c *Client) callOnce(ctx context.Context, promptText string) (*provider.AnalysisResult, error) {
	rawText, err := c.callAPI(ctx, promptText, 0.1, 512)
	if err != nil {
		return nil, err
	}
	parsed, err := parser.Parse(rawText)
	if err != nil {
		return nil, err
	}
	if err := aivalidator.Validate(parsed); err != nil {
		return nil, err
	}
	return &provider.AnalysisResult{
		Category: parsed.Category, Priority: parsed.Priority,
		Sentiment: parsed.Sentiment, RecommendedTeam: parsed.RecommendedTeam,
		Confidence: parsed.Confidence, Summary: parsed.Summary, Tags: parsed.Tags,
	}, nil
}

func (c *Client) callAPI(ctx context.Context, promptText string, temperature float64, maxTokens int) (string, error) {
	start := time.Now()
	reqBody := groqRequest{
		Model:       c.model,
		Messages:    []groqMessage{{Role: "user", Content: promptText}},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq API HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var gr groqResponse
	if err := json.Unmarshal(respBytes, &gr); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if gr.Error != nil {
		return "", fmt.Errorf("groq error: %s", gr.Error.Message)
	}
	if len(gr.Choices) == 0 {
		return "", fmt.Errorf("groq: empty choices")
	}

	utils.Logger.WithField("latency_ms", time.Since(start).Milliseconds()).
		WithField("model", c.model).
		WithField("tokens", gr.Usage.TotalTokens).
		Info("AI: Groq response received")

	return gr.Choices[0].Message.Content, nil
}
