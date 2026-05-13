package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

const path61 = "/gate/METRO-PHM/api/faultRecordsSubsystem/saveRecord"

// Handle61Alarm processes a signal-alarm message: diffs against active state,
// then POSTs start/end records to the platform.
func Handle61Alarm(ctx context.Context, client *PlatformClient, tracker *AlarmTracker, cfg Config, data []byte) {
	var msg SubEventMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WARN] 6.1 alarm: bad json: %v", err)
		return
	}

	var hits []AlarmHit
	if err := json.Unmarshal(msg.Hits, &hits); err != nil || len(hits) == 0 {
		return
	}

	codes := make([]HitCode, len(hits))
	for i, h := range hits {
		codes[i] = HitCode{Code: h.Code, Name: h.Name}
	}

	ts := nowMs()
	diff := tracker.Diff(msg.EventMeta.DeviceID, codes, ts)

	var records []Record61
	for _, hit := range diff.Added {
		records = append(records, buildRecord61(cfg, msg.EventMeta, "0", hit.Code, hit.Name, hit.UUID, hit.StartTime, 0))
	}
	for _, hit := range diff.Removed {
		records = append(records, buildRecord61(cfg, msg.EventMeta, "0", hit.Code, hit.Name, hit.UUID, hit.StartTime, hit.EndTime))
	}

	if len(records) == 0 {
		return
	}
	if err := client.PostJSON(ctx, path61, records); err != nil {
		log.Printf("[ERROR] 6.1 alarm POST failed: %v", err)
	}
}

// Handle61Predict processes a signal-predict message the same way as alarm.
func Handle61Predict(ctx context.Context, client *PlatformClient, tracker *AlarmTracker, cfg Config, data []byte) {
	var msg SubEventMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WARN] 6.1 predict: bad json: %v", err)
		return
	}

	var hits []PredictHit
	if err := json.Unmarshal(msg.Hits, &hits); err != nil || len(hits) == 0 {
		return
	}

	// Prefix device key with "predict:" to keep alarm and predict state tables separate.
	deviceKey := "predict:" + msg.EventMeta.DeviceID
	codes := make([]HitCode, len(hits))
	for i, h := range hits {
		codes[i] = HitCode{Code: h.Code, Name: h.Name}
	}

	ts := nowMs()
	diff := tracker.Diff(deviceKey, codes, ts)

	var records []Record61
	for _, hit := range diff.Added {
		records = append(records, buildRecord61(cfg, msg.EventMeta, "1", hit.Code, hit.Name, hit.UUID, hit.StartTime, 0))
	}
	for _, hit := range diff.Removed {
		records = append(records, buildRecord61(cfg, msg.EventMeta, "1", hit.Code, hit.Name, hit.UUID, hit.StartTime, hit.EndTime))
	}

	if len(records) == 0 {
		return
	}
	if err := client.PostJSON(ctx, path61, records); err != nil {
		log.Printf("[ERROR] 6.1 predict POST failed: %v", err)
	}
}

// buildRecord61 constructs a single 6.1 record.
// endTimeMs == 0 means alarm is still open (endtime = "").
func buildRecord61(cfg Config, meta EventMeta, msgType, code, location, uuid string, startMs, endMs int64) Record61 {
	endTime := ""
	if endMs > 0 {
		endTime = strconv.FormatInt(endMs, 10)
	}

	// Use "KTA" as fallback code for raw fault bits that have no PHM code mapping.
	if msgType == "0" && code == "" {
		code = "KTA"
	}

	return Record61{
		ID:          uuid,
		MessageType: msgType,
		TrainType:   cfg.TrainType,
		TrainNo:     meta.TrainID,
		Coach:       fmt.Sprintf("%s%d", cfg.TrainType, meta.CarriageID),
		Location:    location,
		Code:        code,
		Station1:    "",
		Station2:    "",
		StartTime:   strconv.FormatInt(startMs, 10),
		EndTime:     endTime,
		Subsystem:   cfg.SubsystemCode,
		LineName:    meta.LineID,
	}
}
