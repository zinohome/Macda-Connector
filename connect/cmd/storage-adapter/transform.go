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

func parseTimeText(text string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-1-2 15:4:5",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}
