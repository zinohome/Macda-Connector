package main

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

const path67 = "/gate/METRO-PHM/api/devices/status/train/saveOrUpdate"

// Handle67LifeAction processes a signal-life message and immediately POSTs
// each LifeHit as a per-action life record to the platform.
func Handle67LifeAction(ctx context.Context, client *PlatformClient, cfg Config, data []byte) {
	var msg SubEventMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WARN] 6.7 life action: bad json: %v", err)
		return
	}

	var hits []LifeHit
	if err := json.Unmarshal(msg.Hits, &hits); err != nil || len(hits) == 0 {
		return
	}

	ts := nowMs()
	records := make([]LifeRecord67, 0, len(hits))
	for _, hit := range hits {
		records = append(records, LifeRecord67{
			LineName:     msg.EventMeta.LineID,
			TrainType:    cfg.TrainType,
			TrainNo:      msg.EventMeta.TrainID,
			PartCode:     hit.Code,
			ServiceTime:  ts,
			ServiceValue: hit.Value,
			Mileage:      0,
			UseTime:      0,
			Flag:         2, // absolute value; platform must not accumulate
		})
	}

	if err := client.PostJSON(ctx, path67, records); err != nil {
		log.Printf("[ERROR] 6.7 life action POST failed: %v", err)
	}
}

// Run67DailyBatch fires once per day at DailyBatchHour (default: midnight),
// snapshots the full life cache, and POSTs all known part values to the platform.
func Run67DailyBatch(ctx context.Context, client *PlatformClient, cache *LifeCache, cfg Config) {
	for {
		next := nextOccurrence(cfg.DailyBatchHour)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(next)):
		}

		sendDailyBatch(ctx, client, cache, cfg)

		// Sleep a bit to avoid firing twice if the system clock jitters at the boundary.
		time.Sleep(30 * time.Second)
	}
}

func sendDailyBatch(ctx context.Context, client *PlatformClient, cache *LifeCache, cfg Config) {
	entries := cache.Snapshot()
	if len(entries) == 0 {
		log.Printf("[INFO] 6.7 daily batch: cache empty, skipping")
		return
	}

	ts := nowMs()
	records := make([]LifeRecord67, 0, len(entries))
	for _, e := range entries {
		records = append(records, LifeRecord67{
			LineName:     e.LineName,
			TrainType:    cfg.TrainType,
			TrainNo:      e.TrainID,
			PartCode:     e.PartCode,
			ServiceTime:  ts,
			ServiceValue: e.ServiceValue,
			Mileage:      0,
			UseTime:      0,
			Flag:         2,
		})
	}

	if err := client.PostJSON(ctx, path67, records); err != nil {
		log.Printf("[ERROR] 6.7 daily batch POST failed: %v", err)
		return
	}
	log.Printf("[INFO] 6.7 daily batch sent: %d records", len(records))
}

// nextOccurrence returns the next wall-clock time when the given hour (0-23) occurs.
func nextOccurrence(hour int) time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
