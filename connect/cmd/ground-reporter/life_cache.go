package main

import (
	"fmt"
	"sync"
	"time"
)

// partSpec describes one life-tracked part per carriage.
// offset matches the formula: partCode = carriageID*1000 + 50000 + offset
// unit: "hours" (fan/compressor raw value is seconds) or "count" (valve operation count)
type partSpec struct {
	Offset   int
	Name     string
	RawField string // key in ParsedMsg.Raw
	Unit     string // "hours" | "count"
}

// partSpecs mirrors buildLifeHits in nb67_event_processor.go exactly.
var partSpecs = []partSpec{
	{1, "机组1通风机累计运行时间", "DwefOpTmU11", "hours"},
	{2, "机组1冷凝风机累计运行时间", "DwcfOpTmU11", "hours"},
	{3, "机组1压缩机1累计运行时间", "DwcpOpTmU11", "hours"},
	{4, "机组1压缩机2累计运行时间", "DwcpOpTmU12", "hours"},
	{5, "机组1新风阀开关总次数", "DwfadOpCntU1", "count"},
	{6, "机组1回风阀开关总次数", "DwradOpCntU1", "count"},
	{11, "机组2通风机累计运行时间", "DwefOpTmU21", "hours"},
	{12, "机组2冷凝风机累计运行时间", "DwcfOpTmU21", "hours"},
	{13, "机组2压缩机1累计运行时间", "DwcpOpTmU21", "hours"},
	{14, "机组2压缩机2累计运行时间", "DwcpOpTmU22", "hours"},
	{15, "机组2新风阀开关总次数", "DwfadOpCntU2", "count"},
	{16, "机组2回风阀开关总次数", "DwradOpCntU2", "count"},
	{21, "废排风机累计运行时间", "DwexufanOpTm", "hours"},
	{22, "废排风阀开关总次数", "DwdmpexuOpCnt", "count"},
}

type partCacheKey struct {
	DeviceID string
	PartCode string
}

type PartEntry struct {
	LineName   string
	TrainType  string
	TrainID    string
	CarriageID int
	PartCode   string
	// serviceValue to report: already converted (hours for time-based, count for valve)
	ServiceValue int64
	UpdatedAt    time.Time
}

// LifeCache stores the latest cumulative part values seen in signal-parsed.
// Updated on every parsed message; snapshot is taken for the nightly batch.
type LifeCache struct {
	mu    sync.RWMutex
	cache map[partCacheKey]*PartEntry
}

func newLifeCache() *LifeCache {
	return &LifeCache{cache: make(map[partCacheKey]*PartEntry)}
}

// Update extracts all 14 part values from a signal-parsed message and refreshes the cache.
func (c *LifeCache) Update(msg ParsedMsg, trainType string) {
	if len(msg.Raw) == 0 {
		return
	}
	base := int64(msg.CarriageID*1000 + 50_000)
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, spec := range partSpecs {
		raw := rawInt(msg.Raw, spec.RawField)
		if raw == 0 {
			continue // field absent or device not yet reporting
		}

		var serviceValue int64
		if spec.Unit == "hours" {
			serviceValue = raw / 3600 // device reports seconds; platform expects hours
		} else {
			serviceValue = raw
		}

		code := fmt.Sprintf("%d", base+int64(spec.Offset))
		key := partCacheKey{DeviceID: msg.DeviceID, PartCode: code}

		c.cache[key] = &PartEntry{
			LineName:     msg.LineID,
			TrainType:    trainType,
			TrainID:      msg.TrainID,
			CarriageID:   msg.CarriageID,
			PartCode:     code,
			ServiceValue: serviceValue,
			UpdatedAt:    now,
		}
	}
}

// Snapshot returns a point-in-time copy of all cached entries for the daily batch.
func (c *LifeCache) Snapshot() []PartEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]PartEntry, 0, len(c.cache))
	for _, e := range c.cache {
		out = append(out, *e)
	}
	return out
}
