package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
)

// carriageNames maps carriage ID (1-6) to the physical coach designation.
// Order matches the spec: "Tc1,Mp1,M1,M2,Mp2,Tc2 分别对应 1-6 车厢".
var carriageNames = map[int]string{
	1: "Tc1",
	2: "Mp1",
	3: "M1",
	4: "M2",
	5: "Mp2",
	6: "Tc2",
}

func coachName(carriageID int) string {
	if name, ok := carriageNames[carriageID]; ok {
		return name
	}
	return strconv.Itoa(carriageID)
}

// Handle61Alarm processes a signal-alarm message: diffs against active state,
// then POSTs start/end records to the platform.
func Handle61Alarm(ctx context.Context, client *PlatformClient, tracker *AlarmTracker, sc *StationCache, cfg Config, data []byte) {
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
	si := sc.Get(msg.EventMeta.DeviceID)

	var records []Record61
	for _, hit := range diff.Added {
		records = append(records, buildRecord61(cfg, msg.EventMeta, si, "0", hit.Code, hit.Name, hit.UUID, hit.StartTime, 0))
	}
	for _, hit := range diff.Removed {
		records = append(records, buildRecord61(cfg, msg.EventMeta, si, "0", hit.Code, hit.Name, hit.UUID, hit.StartTime, hit.EndTime))
	}

	if len(records) == 0 {
		return
	}
	if err := client.PostJSON(ctx, cfg.FaultRecordURL, records); err != nil {
		log.Printf("[ERROR] 6.1 alarm POST failed: %v", err)
	}
}

// Handle61Predict processes a signal-predict message the same way as alarm.
func Handle61Predict(ctx context.Context, client *PlatformClient, tracker *AlarmTracker, sc *StationCache, cfg Config, data []byte) {
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
	si := sc.Get(msg.EventMeta.DeviceID)

	var records []Record61
	for _, hit := range diff.Added {
		records = append(records, buildRecord61(cfg, msg.EventMeta, si, "1", hit.Code, hit.Name, hit.UUID, hit.StartTime, 0))
	}
	for _, hit := range diff.Removed {
		records = append(records, buildRecord61(cfg, msg.EventMeta, si, "1", hit.Code, hit.Name, hit.UUID, hit.StartTime, hit.EndTime))
	}

	if len(records) == 0 {
		return
	}
	if err := client.PostJSON(ctx, cfg.FaultRecordURL, records); err != nil {
		log.Printf("[ERROR] 6.1 predict POST failed: %v", err)
	}
}

// buildRecord61 constructs a single 6.1 record.
// endTimeMs == 0 means alarm is still open (endtime = "").
func buildRecord61(cfg Config, meta EventMeta, si StationInfo, msgType, code, location, uuid string, startMs, endMs int64) Record61 {
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
		Coach:       coachName(meta.CarriageID),
		Location:    location,
		Code:        code,
		Station1:    strconv.Itoa(int(si.CurStation)),
		Station2:    strconv.Itoa(int(si.NextStation)),
		StartTime:   strconv.FormatInt(startMs, 10),
		EndTime:     endTime,
		Subsystem:   cfg.SubsystemCode,
		LineName:    meta.LineID,
	}
}
