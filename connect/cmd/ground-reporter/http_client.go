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
	httpClient *http.Client
	retryMax   int
	backoffMs  int
}

func newPlatformClient(cfg Config) *PlatformClient {
	return &PlatformClient{
		apiKey: cfg.PlatformApiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.PlatformTimeoutSec) * time.Second,
		},
		retryMax:  cfg.PlatformRetryMax,
		backoffMs: cfg.PlatformRetryBackoffMs,
	}
}

// PostJSON marshals body as JSON and POSTs to the given full URL, retrying on transient failures.
func (c *PlatformClient) PostJSON(ctx context.Context, url string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	backoff := time.Duration(c.backoffMs) * time.Millisecond

	for attempt := 0; attempt <= c.retryMax; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		if err = c.doPost(ctx, url, data); err == nil {
			return nil
		}
		log.Printf("[WARN] POST %s attempt %d/%d failed: %v", url, attempt+1, c.retryMax+1, err)
	}
	return fmt.Errorf("POST %s: all %d attempts failed, last error: %w", url, c.retryMax+1, err)
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
