package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"gatekeeper/internal/config"
)

// ErrAirGapped is returned when the privacy policy disallows transmitting
// code to a cloud LLM endpoint (privacy.allow_public_cloud_transmission=false).
var ErrAirGapped = errors.New("llm transmission disallowed by air-gap policy (privacy.allow_public_cloud_transmission=false)")

// PillarAdjustments holds LLM-suggested deductions per pillar, matching the
// structured output shape in AGENTS.md.
type PillarAdjustments struct {
	StaticHealthDeduction float64 `json:"static_health_deduction"`
	ArchitectureDeduction float64 `json:"architecture_deduction"`
}

// Remediation is a single LLM-suggested remediation.
type Remediation struct {
	Priority      string `json:"priority"`
	Pillar        string `json:"pillar"`
	Location      string `json:"location"`
	Finding       string `json:"finding"`
	ActionableFix string `json:"actionable_fix"`
}

// Evaluation is the structured output contract for LLM code evaluation.
type Evaluation struct {
	PillarAdjustments PillarAdjustments `json:"pillar_adjustments"`
	Remediations      []Remediation     `json:"remediations"`
}

// systemPrompt instructs the model to respond with only the structured JSON.
const systemPrompt = `You are Gatekeeper, an organizational code quality evaluator.
Analyze the provided code and respond with ONLY a JSON object — no markdown, no prose — matching exactly this shape:
{"pillar_adjustments":{"static_health_deduction":<number 0-20>,"architecture_deduction":<number 0-25>},"remediations":[{"priority":"LOW|MEDIUM|HIGH","pillar":"static|architecture|verification|security","location":"file:lines","finding":"short description of the issue","actionable_fix":"concrete fix"}]}
Deductions reflect quality issues that static analysis cannot catch (readability, intent, duplication, design). Use zero deductions when the code is clean. Keep remediations to the most impactful issues.`

// Client sends code to an OpenAI-compatible LLM endpoint and parses the
// structured evaluation response.
type Client struct {
	cfg *Config
}

// NewClient creates a Client from an LLM Config (see FromConfig).
func NewClient(cfg *Config) *Client {
	return &Client{cfg: cfg}
}

// Evaluate transmits the given code to the configured LLM and returns the
// parsed evaluation. It applies secret scrubbing per the privacy policy,
// refuses to transmit when the air-gap policy is set, and retries transient
// failures up to MaxRetries() additional times before returning an error so
// callers can fall back to rule-based scoring (Story G-2).
func (c *Client) Evaluate(ctx context.Context, privacy config.PrivacyConfig, code string) (*Evaluation, error) {
	if privacy.AllowPublicCloudTransmission != nil && !*privacy.AllowPublicCloudTransmission {
		return nil, ErrAirGapped
	}

	token, err := c.cfg.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("obtain LLM credential: %w", err)
	}

	// Data scrubbing is on by default (spec §5.3).
	if privacy.DataScrubbing == nil || *privacy.DataScrubbing {
		code = ScrubSecrets(code)
	}

	timeout := time.Duration(c.cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 4000 * time.Millisecond
	}
	temp := float32(c.cfg.Temperature)

	var lastErr error
	for attempt := 0; attempt <= MaxRetries(); attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)
		eval, err := c.attempt(attemptCtx, token, temp, code, timeout)
		cancel()
		if err == nil {
			return eval, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break // caller cancelled; do not burn remaining retries
		}
	}
	return nil, fmt.Errorf("llm evaluation failed after %d attempts: %w", MaxRetries()+1, lastErr)
}

// attempt performs a single chat completion request and parses its response.
// The timeout is enforced at the HTTP client level because the underlying
// component does not reliably honour context cancellation on in-flight requests.
func (c *Client) attempt(ctx context.Context, token string, temp float32, code string, timeout time.Duration) (*Evaluation, error) {
	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     c.cfg.BaseURL,
		APIKey:      token,
		Model:       c.cfg.ModelName,
		Timeout:     timeout,
		Temperature: &temp,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	resp, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(code),
	})
	if err != nil {
		return nil, fmt.Errorf("chat completion request: %w", err)
	}

	var eval Evaluation
	if err := json.Unmarshal([]byte(resp.Content), &eval); err != nil {
		return nil, fmt.Errorf("parse structured evaluation from model output: %w", err)
	}
	return &eval, nil
}
