package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const compressPrompt = `You are a D&D campaign archivist. Compress the following session events into
a 150-word summary that preserves:

- Major plot developments and turning points
- Character decisions with lasting consequences
- New NPCs introduced and their story significance
- Changes to world state (what is now permanently different)

Rules:
- Be specific. Use character names.
- Omit blow-by-blow combat unless the outcome changed the story.
- Preserve emotional beats and character moments.
- Output only the summary - no preamble, no metadata.
- End with a section: "OPEN HOOKS:" followed by bullet list of
  unresolved threads that should carry into the next session.

SESSION EVENTS:
%s`

// Compressor summarizes raw session events, optionally through Anthropic.
type Compressor struct {
	apiKey string
	client *http.Client
}

// NewCompressor creates a compressor backed by Anthropic API when key is present.
func NewCompressor(apiKey string) *Compressor {
	return &Compressor{
		apiKey: strings.TrimSpace(apiKey),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Compress returns an AI-compressed summary or deterministic fallback if API is unavailable.
func (c *Compressor) Compress(ctx context.Context, rawEvents string) (string, error) {
	rawEvents = strings.TrimSpace(rawEvents)
	if rawEvents == "" {
		return "", fmt.Errorf("raw_events is required")
	}

	if c.apiKey == "" {
		return fallbackCompress(rawEvents), nil
	}

	summary, err := c.compressAnthropic(ctx, rawEvents)
	if err != nil {
		return fallbackCompress(rawEvents), nil
	}
	return summary, nil
}

func (c *Compressor) compressAnthropic(ctx context.Context, rawEvents string) (string, error) {
	payload := map[string]interface{}{
		"model":       "claude-3-5-sonnet-latest",
		"max_tokens":  500,
		"temperature": 0.2,
		"messages": []map[string]string{
			{"role": "user", "content": fmt.Sprintf(compressPrompt, rawEvents)},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal anthropic payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create anthropic request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("anthropic returned status %d", resp.StatusCode)
	}

	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decode anthropic response: %w", err)
	}

	parts := make([]string, 0, len(response.Content))
	for _, part := range response.Content {
		if part.Type == "text" {
			parts = append(parts, strings.TrimSpace(part.Text))
		}
	}
	summary := strings.TrimSpace(strings.Join(parts, "\n"))
	if summary == "" {
		return "", fmt.Errorf("anthropic returned empty summary")
	}
	return summary, nil
}

func fallbackCompress(rawEvents string) string {
	clean := strings.Join(strings.Fields(rawEvents), " ")
	if len(clean) > 750 {
		clean = clean[:750] + "..."
	}

	return strings.TrimSpace(clean + "\n\nOPEN HOOKS:\n- Review unresolved plot threads from this session.")
}
