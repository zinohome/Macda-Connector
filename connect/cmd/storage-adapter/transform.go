package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func (r *StorageRecord) toInsertArgs(rawValue []byte) ([]any, error) {
	now := time.Now().UTC()
	eventTime, eventTimeValid := parseEventTime(r.EventTimeText, r.EventTimeValid, now)
	ingestTime := parseOptionalTimeOrDefault(r.IngestTime, now)
	processTime := parseOptionalTime(r.ProcessTime)

	if len(r.PayloadJSON) == 0 {
		r.PayloadJSON = append(r.PayloadJSON, rawValue...)
	}
	if !json.Valid(r.PayloadJSON) {
		return nil, fmt.Errorf("payload_json is invalid json")
	}

	return []any{
		eventTime,
		ingestTime,
		processTime,
		r.LineID,
		r.TrainID,
		r.CarriageID,
		r.DeviceID,
		r.FrameNo,
		r.ParserVersion,
		r.QualityCode,
		eventTimeValid,
		r.WmodeU1,
		r.WmodeU2,
		r.FCpU11,
		r.FCpU12,
		r.FCpU21,
		r.FCpU22,
		r.ICpU11,
		r.ICpU12,
		r.ICpU21,
		r.ICpU22,
		r.SuckpU11,
		r.SuckpU12,
		r.SuckpU21,
		r.SuckpU22,
		r.HighpressU11,
		r.HighpressU12,
		r.HighpressU21,
		r.HighpressU22,
		r.FasU1,
		r.FasU2,
		r.RasU1,
		r.RasU2,
		r.PresdiffU1,
		r.PresdiffU2,
		r.AqCo2U1,
		r.AqCo2U2,
		r.AqTvocU1,
		r.AqTvocU2,
		r.AqPm25U1,
		r.AqPm25U2,
		r.AqPm10U1,
		r.AqPm10U2,
		r.BocfltEfU11,
		r.BocfltEfU12,
		r.BocfltEfU21,
		r.BocfltEfU22,
		r.BocfltCfU11,
		r.BocfltCfU12,
		r.BocfltCfU21,
		r.BocfltCfU22,
		r.BlpfltCompU11,
		r.BlpfltCompU12,
		r.BlpfltCompU21,
		r.BlpfltCompU22,
		r.BscfltCompU11,
		r.BscfltCompU12,
		r.BscfltCompU21,
		r.BscfltCompU22,
		r.BfltTempover,
		r.BfltDiffpresU1,
		r.BfltDiffpresU2,
		r.PayloadJSON,
	}, nil
}

func (r *EventFlatRecord) toEventInsertArgs() ([]any, error) {
	now := time.Now().UTC()
	eventTime, _ := parseEventTime(r.EventTime, true, now)
	ingestTime := parseOptionalTimeOrDefault(r.IngestTime, now)

	return []any{
		eventTime,
		ingestTime,
		r.LineID,
		r.TrainID,
		r.CarriageID,
		r.DeviceID,
		r.EventType,
		r.FaultCode,
		r.FaultName,
		r.Severity,
		r.Payload,
	}, nil
}

func parseEventTime(text string, valid bool, fallback time.Time) (time.Time, bool) {
	if text == "" {
		return fallback, false
	}
	if parsed, ok := parseTimeText(text); ok {
		if valid {
			return parsed, true
		}
		return parsed, false
	}
	return fallback, false
}

func parseOptionalTime(text string) *time.Time {
	if text == "" {
		return nil
	}
	if parsed, ok := parseTimeText(text); ok {
		return &parsed
	}
	return nil
}

func parseOptionalTimeOrDefault(text string, fallback time.Time) time.Time {
	if parsed, ok := parseTimeText(text); ok {
		return parsed
	}
	return fallback
}

// cnLoc 是 Asia/Shanghai 时区。上游 event_time_text 是裸字符串（无时区，实为
// 现场本地时间），用 time.Parse 会被默认按 UTC 解读 → 入库 TIMESTAMPTZ 比真实时刻 +8h，
// 即 GitHub #24 / RET-46 报告的根因。修复：裸 layout 必须用 ParseInLocation。
var cnLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil || loc == nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}()

func parseTimeText(text string) (time.Time, bool) {
	// 带时区 layout：必须用 time.Parse，否则 +08:00/Z 会被 ParseInLocation 二次本地化。
	tzLayouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range tzLayouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC(), true
		}
	}
	// 裸 layout：按 Asia/Shanghai 解析，再归一化为 UTC。
	nakedLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-1-2 15:4:5",
		"2006-01-02T15:04:05",
	}
	for _, layout := range nakedLayouts {
		if parsed, err := time.ParseInLocation(layout, text, cnLoc); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}
