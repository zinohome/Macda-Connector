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

// ============================================================
// Phase 5: 制冷预警规则回归测试（对应 GitHub #21 现象 A）
// 期望：
//   - 压缩机不运行（FCpUx1=0 或 FCpUx2=0）→ 永不触发
//   - 两台压缩机频率不同 → 永不触发
//   - 同频率运行 + 电流差 > 2A，持续时间不足 → 不触发
//   - 同频率运行 + 电流差 > 2A，持续 ≥ DurationSeconds → 触发 HVAC{base+5}
// ============================================================

func TestCoolingSystem_NotTriggeredWhenCompressorStopped(t *testing.T) {
	orig := globalConfigStore
	t.Cleanup(func() { globalConfigStore = orig })

	cfg := configMap{
		"WARN_COOLING_SYSTEM": {
			TriggerValue: 20, ClearValue: 20, DurationSeconds: 180, // 3min
			Enabled: true, RawScale: 1,
		},
	}
	cs := &ConfigStore{}
	cs.val.Store(&cfg)
	globalConfigStore = cs

	p := &NB67EventProcessor{
		logger:  service.MockResources().Logger(),
		runtime: "PRD",
	}

	// 机组 1：压缩机 2 未运行（FCpU12=0），FCpU11=50 → f1 != f2 不应触发
	raw := map[string]any{
		"FCpU11": int64(50), "FCpU12": int64(0),
		"ICpU11": int64(100), "ICpU12": int64(20), // 电流差 8A
	}
	now := time.Now()
	for i := 0; i < 5; i++ {
		hits := p.buildPredictHits(raw, 1, "dev-cool-stop", now.Add(time.Duration(i)*time.Minute))
		if hasPredictCode(hits, "HVAC105") {
			t.Fatalf("HVAC105 must NOT trigger when compressor not running (frame %d)", i)
		}
	}
}

func TestCoolingSystem_NotTriggeredWhenDifferentFrequency(t *testing.T) {
	orig := globalConfigStore
	t.Cleanup(func() { globalConfigStore = orig })

	cfg := configMap{
		"WARN_COOLING_SYSTEM": {
			TriggerValue: 20, ClearValue: 20, DurationSeconds: 180,
			Enabled: true, RawScale: 1,
		},
	}
	cs := &ConfigStore{}
	cs.val.Store(&cfg)
	globalConfigStore = cs

	p := &NB67EventProcessor{
		logger:  service.MockResources().Logger(),
		runtime: "PRD",
	}

	// 两台压缩机频率不同 (50 vs 40) → 永不触发，即使电流差很大
	raw := map[string]any{
		"FCpU11": int64(50), "FCpU12": int64(40),
		"ICpU11": int64(150), "ICpU12": int64(20),
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		hits := p.buildPredictHits(raw, 1, "dev-cool-diff-freq", now.Add(time.Duration(i)*time.Minute))
		if hasPredictCode(hits, "HVAC105") {
			t.Fatalf("HVAC105 must NOT trigger when frequencies differ (frame %d)", i)
		}
	}
}

func TestCoolingSystem_TriggersAfterDurationOnSameFrequency(t *testing.T) {
	orig := globalConfigStore
	t.Cleanup(func() { globalConfigStore = orig })

	cfg := configMap{
		"WARN_COOLING_SYSTEM": {
			TriggerValue: 20, ClearValue: 20, DurationSeconds: 180,
			Enabled: true, RawScale: 1,
		},
	}
	cs := &ConfigStore{}
	cs.val.Store(&cfg)
	globalConfigStore = cs

	p := &NB67EventProcessor{
		logger:  service.MockResources().Logger(),
		runtime: "PRD",
	}

	// 同频 50Hz 运行，电流差 30A > 2A 阈值
	raw := map[string]any{
		"FCpU11": int64(50), "FCpU12": int64(50),
		"ICpU11": int64(150), "ICpU12": int64(120), // 差 30 > 阈 20
	}
	now := time.Now()

	// 持续时间不足 → 不触发
	first := p.buildPredictHits(raw, 1, "dev-cool-trigger", now)
	if hasPredictCode(first, "HVAC105") {
		t.Fatalf("HVAC105 should not trigger before duration elapses")
	}

	// 4 分钟后 → 触发
	later := p.buildPredictHits(raw, 1, "dev-cool-trigger", now.Add(4*time.Minute))
	if !hasPredictCode(later, "HVAC105") {
		t.Fatalf("HVAC105 should trigger after 3min duration on same-frequency current diff")
	}
}
