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
	base       string // e.g. "http://10.12.48.187:8188"
	httpClient *http.Client
	retryMax   int
	backoffMs  int
}

func newPlatformClient(cfg Config) *PlatformClient {
	return &PlatformClient{
		base: fmt.Sprintf("http://%s:%d", cfg.PlatformIP, cfg.PlatformPort),
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.PlatformTimeoutSec) * time.Second,
		},
		retryMax:  cfg.PlatformRetryMax,
		backoffMs: cfg.PlatformRetryBackoffMs,
	}
}

// PostJSON marshals body as JSON array and POSTs to path, retrying on transient failures.
// body may be a single struct or a slice; it is always wrapped in [] when it's not already a slice.
func (c *PlatformClient) PostJSON(ctx context.Context, path string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := c.base + path
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
		log.Printf("[WARN] POST %s attempt %d/%d failed: %v", path, attempt+1, c.retryMax+1, err)
	}
	return fmt.Errorf("POST %s: all %d attempts failed, last error: %w", path, c.retryMax+1, err)
}

func (c *PlatformClient) doPost(ctx context.Context, url string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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
