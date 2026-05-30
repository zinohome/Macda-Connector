package main

import (
	"testing"
	"time"

	"github.com/benthosdev/benthos/v4/public/service"
)

func TestBuildPredictHits_UsesDurationByWarnCode(t *testing.T) {
	orig := globalConfigStore
	t.Cleanup(func() {
		globalConfigStore = orig
	})

	cfg := configMap{
		"WARN_EF_CURRENT": {
			TriggerValue:    18,
			ClearValue:      18,
			DurationSeconds: 600, // 10m
			Enabled:         true,
			RawScale:        1,
		},
		"WARN_CF_CURRENT": {
			TriggerValue:    23,
			ClearValue:      23,
			DurationSeconds: 60, // 1m
			Enabled:         true,
			RawScale:        1,
		},
		"WARN_EXUF_CURRENT": {
			TriggerValue:    23,
			ClearValue:      23,
			DurationSeconds: 600, // 10m
			Enabled:         true,
			RawScale:        1,
		},
	}
	cs := &ConfigStore{}
	cs.val.Store(&cfg)
	globalConfigStore = cs

	p := &NB67EventProcessor{
		logger:  service.MockResources().Logger(),
		runtime: "PRD",
	}

	raw := map[string]any{
		"CfbkCfU11": true,
		"ICfU11":    int64(30),
	}
	deviceID := "dev-cf-1"
	now := time.Now()

	first := p.buildPredictHits(raw, 3, deviceID, now)
	if hasPredictCode(first, "HVAC316") {
		t.Fatalf("HVAC316 should not trigger on first frame")
	}

	second := p.buildPredictHits(raw, 3, deviceID, now.Add(2*time.Minute))
	if !hasPredictCode(second, "HVAC316") {
		t.Fatalf("HVAC316 should trigger after WARN_CF_CURRENT duration")
	}
}

func hasPredictCode(hits []PredictHit, code string) bool {
	for _, h := range hits {
		if h.Code == code {
			return true
		}
	}
	return false
}
