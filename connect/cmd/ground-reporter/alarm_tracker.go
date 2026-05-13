package main

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
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

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
