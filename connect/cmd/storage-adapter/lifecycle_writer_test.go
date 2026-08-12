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

// TestApplySweepClosed_RemovesFromMemory 验证 sweeper 关闭 DB 行后，
// 内存 active map 里对应 (device, code) 也被同步删除；
// 否则同一 code 下一帧到来时会被 diff 判为 touch，UPDATE 无匹配行，日志刷屏。
// 这是 issue #26 sweeper 兜底方案的内存一致性保障。
func TestApplySweepClosed_RemovesFromMemory(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}
	// 预置：DEV_A 活跃 HVAC112 + HVAC209，DEV_B 活跃 HVAC301
	w.applyMemDelta("DEV_A", []string{"HVAC112", "HVAC209"}, nil)
	w.applyMemDelta("DEV_B", []string{"HVAC301"}, nil)

	// 模拟 sweeper 从 DB RETURNING 回来的关闭列表：
	//   - DEV_A/HVAC112 被 sweeper close
	//   - DEV_B/HVAC301 被 sweeper close（DEV_B 应彻底从 map 中消失）
	w.applySweepClosed([]SweepClosed{
		{DeviceID: "DEV_A", FaultCode: "HVAC112"},
		{DeviceID: "DEV_B", FaultCode: "HVAC301"},
	})

	// DEV_A 剩 HVAC209
	if _, ok := w.active["DEV_A"]["HVAC112"]; ok {
		t.Errorf("DEV_A/HVAC112 应已从 active map 移除")
	}
	if _, ok := w.active["DEV_A"]["HVAC209"]; !ok {
		t.Errorf("DEV_A/HVAC209 未被 sweeper 关闭，应仍在 active map 中")
	}

	// DEV_B 只有一条 code，全部关掉后 map 里 DEV_B 键应被清理
	if _, ok := w.active["DEV_B"]; ok {
		t.Errorf("DEV_B 的所有 code 都被 close，DEV_B 键应从 active map 删除，got %v", w.active["DEV_B"])
	}
}

// TestApplySweepClosed_ThenNextFrame_TreatsAsOpen 验证 sweeper 关闭后，
// 同 device 同 code 的新一帧应被 diff 判成 OPEN（等价于全新预警），
// 而不是被错判成 TOUCH 打空——这是 sweeper 与 diff 状态机协同的正确性要求。
func TestApplySweepClosed_ThenNextFrame_TreatsAsOpen(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}
	// 帧 1：DEV_C 命中 HVAC112 → open
	toOpen, _, _ := w.computeDiff("DEV_C", []string{"HVAC112"})
	if len(toOpen) != 1 {
		t.Fatalf("frame1 open=%d, want 1", len(toOpen))
	}
	w.applyMemDelta("DEV_C", toOpen, nil)

	// sweeper 兜底 close 掉 DEV_C/HVAC112
	w.applySweepClosed([]SweepClosed{{DeviceID: "DEV_C", FaultCode: "HVAC112"}})

	// 帧 2：DEV_C 再次命中 HVAC112 —— 应视为全新 open，而非 touch
	toOpen, toClose, toTouch := w.computeDiff("DEV_C", []string{"HVAC112"})
	if len(toOpen) != 1 {
		t.Errorf("post-sweep frame open=%d, want 1 (应视为全新预警)", len(toOpen))
	}
	if len(toClose) != 0 || len(toTouch) != 0 {
		t.Errorf("post-sweep frame close/touch = %d/%d, want 0/0", len(toClose), len(toTouch))
	}
}

// TestApplySweepClosed_EmptyInput 保证空输入是 no-op、不 panic。
func TestApplySweepClosed_EmptyInput(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}
	w.applyMemDelta("DEV_D", []string{"HVAC112"}, nil)
	w.applySweepClosed(nil)
	w.applySweepClosed([]SweepClosed{})
	if _, ok := w.active["DEV_D"]["HVAC112"]; !ok {
		t.Errorf("空 sweep 输入不应影响任何 active 状态")
	}
}

// TestApplySweepClosed_UnknownDevice 保证 sweeper 关闭的行如果内存里没有对应记录（例如
// 进程重启后 Recover 之前的窗口），不 panic、不做无意义操作。
func TestApplySweepClosed_UnknownDevice(t *testing.T) {
	w := &LifecycleWriter{active: map[string]map[string]struct{}{}}
	w.applySweepClosed([]SweepClosed{
		{DeviceID: "GHOST_DEV", FaultCode: "HVAC999"},
	})
	if len(w.active) != 0 {
		t.Errorf("未知 device 的 sweep close 不应产生任何 active map 条目, got %v", w.active)
	}
}
