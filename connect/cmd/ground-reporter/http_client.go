package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type PlatformClient struct {
	apiKey     string // X-Api-Key header value; empty = omit header
	xToken     string // x-token header value; empty = omit header
	httpClient *http.Client
	retryMax   int
	backoffMs  int
}

func newPlatformClient(cfg Config) *PlatformClient {
	return &PlatformClient{
		apiKey: cfg.PlatformApiKey,
		xToken: cfg.PlatformXToken,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.PlatformTimeoutSec) * time.Second,
		},
		retryMax:  cfg.PlatformRetryMax,
		backoffMs: cfg.PlatformRetryBackoffMs,
	}
}

// PostJSON marshals body as JSON and POSTs to the given full URL, retrying on transient failures.
// SetEscapeHTML(false) prevents & < > from being escaped to & etc., so platform receives
// location values like "空调机组1&2" verbatim instead of "空调机组1&2".
func (c *PlatformClient) PostJSON(ctx context.Context, url string, body any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(body); err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data := buf.Bytes()
	// json.Encoder.Encode appends a trailing newline; trim it so the body is compact JSON.
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	backoff := time.Duration(c.backoffMs) * time.Millisecond
	var lastErr error

	for attempt := 0; attempt <= c.retryMax; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		if lastErr = c.doPost(ctx, url, data); lastErr == nil {
			return nil
		}
		log.Printf("[WARN] POST %s attempt %d/%d failed: %v", url, attempt+1, c.retryMax+1, lastErr)
	}
	return fmt.Errorf("POST %s: all %d attempts failed, last error: %w", url, c.retryMax+1, lastErr)
}

func (c *PlatformClient) doPost(ctx context.Context, url string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	}
	if c.xToken != "" {
		req.Header.Set("x-token", c.xToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
