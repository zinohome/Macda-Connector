package main

import (
	"testing"
)

// TestInferWarnCode 验证 fault_code → warn_code 的映射与 BFF 端 hvacSeqToWarnCode 对齐。
func TestInferWarnCode(t *testing.T) {
	cases := []struct {
		faultCode string
		want      string
	}{
		{"HVAC301", "WARN_REFRIGERANT_LEAK_COOLING"}, // car 3, seq 1
		{"HVAC301_v", "WARN_REFRIGERANT_LEAK_VENT"},
		{"HVAC305", "WARN_COOLING_SYSTEM"},
		{"HVAC112", "WARN_EF_CURRENT"}, // car 1, seq 12
		{"HVAC125", "WARN_AQ_CO2"},
		{"HVAC120", "WARN_EXUF_CURRENT"},
		{"HVAC207", "WARN_TEMP_SENSOR"},
		{"HVAC209", "WARN_CABIN_OVERHEAT"},
		{"NOT_HVAC", ""},
	}
	for _, c := range cases {
		got := inferWarnCode(c.faultCode)
		if got != c.want {
			t.Errorf("inferWarnCode(%q) = %q, want %q", c.faultCode, got, c.want)
		}
	}
}

// TestInferUnitID 验证 fault_code → unit_id 推断（单数侧→机组1，双数侧→机组2）。
func TestInferUnitID(t *testing.T) {
	cases := []struct {
		faultCode string
		want      int16
		isNil     bool
	}{
		{"HVAC112", 1, false}, // 通风机1电流预警
		{"HVAC215", 2, false}, // 通风机2电流预警
		{"HVAC120", 0, true},  // 废排（公共）
		{"HVAC301", 1, false}, // 冷媒1系统1
		{"HVAC304", 2, false}, // 冷媒2系统2
		{"NOT_HVAC", 0, true},
	}
	for _, c := range cases {
		got := inferUnitID(c.faultCode)
		if c.isNil {
			if got != nil {
				t.Errorf("inferUnitID(%q) = %d, want nil", c.faultCode, *got)
			}
			continue
		}
		if got == nil {
			t.Errorf("inferUnitID(%q) = nil, want %d", c.faultCode, c.want)
			continue
		}
		if *got != c.want {
			t.Errorf("inferUnitID(%q) = %d, want %d", c.faultCode, *got, c.want)
		}
	}
}

// TestLifecycleDiff_ContinuousFrames 验证连续多帧同一 fault_code 命中只算 1 次 open。
// 这是修复 "一条预警写 N 行" 的核心保障：状态机判定 open 只在新出现时触发。
func TestLifecycleDiff_ContinuousFrames(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}

	// 模拟同 device 连续 100 帧都命中 HVAC112
	deviceID := "DEV_A"
	totalOpen := 0
	totalClose := 0
	totalTouch := 0
	for i := 0; i < 100; i++ {
		toOpen, toClose, toTouch := w.computeDiff(deviceID, []string{"HVAC112"})
		totalOpen += len(toOpen)
		totalClose += len(toClose)
		totalTouch += len(toTouch)
		w.applyMemDelta(deviceID, toOpen, toClose)
	}
	if totalOpen != 1 {
		t.Errorf("expected 1 open across 100 same-code frames, got %d", totalOpen)
	}
	if totalClose != 0 {
		t.Errorf("expected 0 close, got %d", totalClose)
	}
	if totalTouch != 99 {
		t.Errorf("expected 99 touch (frames 2-100), got %d", totalTouch)
	}
}

// TestLifecycleDiff_HitToEmptyFrame 验证：命中→空帧 = 1 次 close。
func TestLifecycleDiff_HitToEmptyFrame(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}
	deviceID := "DEV_B"

	// 帧 1：命中
	toOpen, toClose, _ := w.computeDiff(deviceID, []string{"HVAC112"})
	if len(toOpen) != 1 || len(toClose) != 0 {
		t.Fatalf("frame1: expected open=1 close=0, got open=%d close=%d", len(toOpen), len(toClose))
	}
	w.applyMemDelta(deviceID, toOpen, toClose)

	// 帧 2：空帧 → close
	toOpen, toClose, _ = w.computeDiff(deviceID, nil)
	if len(toOpen) != 0 || len(toClose) != 1 {
		t.Fatalf("frame2 (empty): expected open=0 close=1, got open=%d close=%d", len(toOpen), len(toClose))
	}
	if toClose[0] != "HVAC112" {
		t.Errorf("close code = %q, want HVAC112", toClose[0])
	}
	w.applyMemDelta(deviceID, toOpen, toClose)

	// 帧 3：空帧再来 → 不应有任何变化
	toOpen, toClose, _ = w.computeDiff(deviceID, nil)
	if len(toOpen) != 0 || len(toClose) != 0 {
		t.Errorf("frame3 idempotent: expected 0/0, got %d/%d", len(toOpen), len(toClose))
	}
}

// TestLifecycleDiff_ConcurrentDifferentDevices 验证 per-device 状态隔离。
func TestLifecycleDiff_ConcurrentDifferentDevices(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}

	for _, d := range []string{"DEV_X", "DEV_Y"} {
		toOpen, _, _ := w.computeDiff(d, []string{"HVAC112"})
		w.applyMemDelta(d, toOpen, nil)
	}

	// DEV_X 清空，DEV_Y 应继续保持
	_, toClose, _ := w.computeDiff("DEV_X", nil)
	if len(toClose) != 1 {
		t.Fatalf("DEV_X close=%d, want 1", len(toClose))
	}
	w.applyMemDelta("DEV_X", nil, toClose)

	// DEV_Y 状态不受影响 — 再来相同帧应是 touch，不是 open
	toOpenY, toCloseY, toTouchY := w.computeDiff("DEV_Y", []string{"HVAC112"})
	if len(toOpenY) != 0 || len(toCloseY) != 0 || len(toTouchY) != 1 {
		t.Errorf("DEV_Y after DEV_X close: open=%d close=%d touch=%d, want 0/0/1",
			len(toOpenY), len(toCloseY), len(toTouchY))
	}
}

// ============================================================
// 测试辅助：把 applyFrame 里的 diff 计算暴露给单测（不涉及 DB）
// ============================================================

func (w *LifecycleWriter) computeDiff(deviceID string, currCodes []string) (toOpen []string, toClose []string, toTouch []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	prev, ok := w.active[deviceID]
	if !ok {
		prev = make(map[string]struct{})
		w.active[deviceID] = prev
	}
	curr := make(map[string]struct{}, len(currCodes))
	for _, c := range currCodes {
		curr[c] = struct{}{}
	}
	for c := range curr {
		if _, exists := prev[c]; exists {
			toTouch = append(toTouch, c)
		} else {
			toOpen = append(toOpen, c)
		}
	}
	for c := range prev {
		if _, stillActive := curr[c]; !stillActive {
			toClose = append(toClose, c)
		}
	}
	return
}

func (w *LifecycleWriter) applyMemDelta(deviceID string, toOpen []string, toClose []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	prev := w.active[deviceID]
	if prev == nil {
		prev = map[string]struct{}{}
		w.active[deviceID] = prev
	}
	for _, c := range toOpen {
		prev[c] = struct{}{}
	}
	for _, c := range toClose {
		delete(prev, c)
	}
}
