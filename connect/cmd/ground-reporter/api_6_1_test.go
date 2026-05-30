package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestHandle61Alarm_DoesNotPostToPlatform(t *testing.T) {
	var postCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := Config{
		FaultRecordURL: srv.URL,
		TrainType:      "T",
		SubsystemCode:  "12",
	}
	client := &PlatformClient{
		httpClient: srv.Client(),
	}
	tracker := newAlarmTracker()
	sc := newStationCache()

	msg := SubEventMsg{
		EventMeta: EventMeta{
			LineID:        "1",
			TrainID:       "101",
			CarriageID:    3,
			DeviceID:      "dev-3",
			EventTimeText: "2026-05-30 16:00:00",
		},
		Hits: mustJSONRawMessage(t, []AlarmHit{
			{Code: "HVAC327", Name: "故障", Level: 1},
		}),
		Source: "test",
	}

	Handle61Alarm(context.Background(), client, tracker, sc, cfg, mustJSONBytes(t, msg))

	if got := postCount.Load(); got != 0 {
		t.Fatalf("expected no platform post for alarm/fault, got %d", got)
	}
}

func mustJSONBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return b
}

func mustJSONRawMessage(t *testing.T, v any) json.RawMessage {
	t.Helper()
	return mustJSONBytes(t, v)
}
