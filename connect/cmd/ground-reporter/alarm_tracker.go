package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type activeAlarm struct {
	UUID      string
	StartTime int64 // unix ms at first detection
}

// AlarmTracker maintains the lifecycle state for 6.1 alarm/predict records.
// For each device, it tracks which hit codes are currently active.
// When a code appears for the first time → fire "start" (endtime empty).
// When a code disappears → fire "end" (endtime = now).
type AlarmTracker struct {
	mu     sync.Mutex
	active map[string]map[string]*activeAlarm // deviceID → code → alarm
}

func newAlarmTracker() *AlarmTracker {
	return &AlarmTracker{
		active: make(map[string]map[string]*activeAlarm),
	}
}

type AlarmDiff struct {
	Added   []ActiveHit // new alarms to report (endtime = "")
	Removed []ActiveHit // ended alarms to report (endtime = now)
}

type ActiveHit struct {
	UUID      string
	Code      string
	Name      string
	StartTime int64 // ms
	EndTime   int64 // ms; 0 means not yet ended
}

// Diff computes which hit codes are new vs. which have ended for the given device.
// currentCodes is the full set of hit codes present in the current message.
// The caller provides nowMs as the reference time so the tracker is deterministic.
func (t *AlarmTracker) Diff(deviceID string, currentCodes []HitCode, nowMs int64) AlarmDiff {
	t.mu.Lock()
	defer t.mu.Unlock()

	existing, ok := t.active[deviceID]
	if !ok {
		existing = make(map[string]*activeAlarm)
		t.active[deviceID] = existing
	}

	currSet := make(map[string]HitCode, len(currentCodes))
	for _, h := range currentCodes {
		currSet[h.Code] = h
	}

	var diff AlarmDiff

	// Codes present now but not before → new alarm
	for code, hit := range currSet {
		if _, exists := existing[code]; !exists {
			uuid := newUUID()
			existing[code] = &activeAlarm{UUID: uuid, StartTime: nowMs}
			diff.Added = append(diff.Added, ActiveHit{
				UUID:      uuid,
				Code:      code,
				Name:      hit.Name,
				StartTime: nowMs,
			})
		}
	}

	// Codes present before but not now → alarm ended
	for code, alarm := range existing {
		if _, stillActive := currSet[code]; !stillActive {
			diff.Removed = append(diff.Removed, ActiveHit{
				UUID:      alarm.UUID,
				Code:      code,
				Name:      "",
				StartTime: alarm.StartTime,
				EndTime:   nowMs,
			})
			delete(existing, code)
		}
	}

	return diff
}

// HitCode carries the minimal info needed for diff (code + display name).
type HitCode struct {
	Code string
	Name string
}

// newUUID generates a 32-char hex UUID (no dashes) matching platform field max-length=32.
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x%04x%04x%04x%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}

// RecoverFromLifecycle 在 ground-reporter 启动时从 hvac.warning_lifecycle 拉活跃集合
// 重建内存 active map，避免重启后把已 active 的预警当作新 open 重发给平台。
//
// 注：恢复的条目 UUID 仍是新生成的（DB 表没存 platform-side UUID），但 Diff 算法
// 只用 UUID 做"同一预警的关联标识"，平台侧用 (device, code) 作为业务键，新 UUID 不会引发重复 open。
// 关键效果：本次启动后再次出现已恢复的 code 时，Diff 看到它"已 active"，不会发 open。
func (t *AlarmTracker) RecoverFromLifecycle(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `
		SELECT device_id, fault_code, start_time
		  FROM hvac.warning_lifecycle
		 WHERE end_time IS NULL
	`)
	if err != nil {
		return fmt.Errorf("alarm tracker recover query: %w", err)
	}
	defer rows.Close()

	t.mu.Lock()
	defer t.mu.Unlock()

	total := 0
	for rows.Next() {
		var deviceID, faultCode string
		var startTime time.Time
		if err := rows.Scan(&deviceID, &faultCode, &startTime); err != nil {
			return fmt.Errorf("alarm tracker recover scan: %w", err)
		}
		// GitHub #23 / RET-46 修复（2026-06-12）：
		// Handle61Predict 用 "predict:" + deviceID 作为 tracker key（见 api_6_1.go），
		// 此处恢复时必须用相同前缀，否则重启后已活跃 predict 不会命中 tracker → 重新
		// 触发 diff.Added → 给平台重发 open 报文（症状 D：重启风暴）。
		key := "predict:" + deviceID
		existing, ok := t.active[key]
		if !ok {
			existing = make(map[string]*activeAlarm)
			t.active[key] = existing
		}
		existing[faultCode] = &activeAlarm{
			UUID:      newUUID(),
			StartTime: startTime.UnixMilli(),
		}
		total++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("alarm tracker recover iter: %w", err)
	}
	log.Printf("[INFO] alarm tracker recover: %d active codes across %d devices", total, len(t.active))
	return nil
}
