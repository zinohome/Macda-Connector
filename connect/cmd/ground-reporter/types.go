package main

import "encoding/json"

// --- Kafka message types (mirror connect-nb67 output) ---

type EventMeta struct {
	LineID        string `json:"line_id"`
	TrainID       string `json:"train_id"`
	CarriageID    int    `json:"carriage_id"`
	DeviceID      string `json:"device_id"`
	EventTimeText string `json:"event_time_text"`
	IngestTime    string `json:"ingest_time"`
}

// SubEventMsg is the top-level message on signal-alarm, signal-predict, signal-life.
type SubEventMsg struct {
	EventMeta EventMeta       `json:"event_meta"`
	Hits      json.RawMessage `json:"hits"`
	Source    string          `json:"source"`
}

type AlarmHit struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type PredictHit struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Severity int    `json:"severity"`
}

type LifeHit struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Severity int    `json:"severity"`
	Value    int64  `json:"value"` // seconds (fans/compressors) or counts (valves)
	Limit    int64  `json:"limit"`
}

// ParsedMsg is a minimal view of signal-parsed used only for life cache population.
type ParsedMsg struct {
	LineID     string         `json:"line_id"`
	TrainID    string         `json:"train_id"`
	CarriageID int            `json:"carriage_id"`
	DeviceID   string         `json:"device_id"`
	IngestTime string         `json:"ingest_time"`
	Raw        map[string]any `json:"raw"`
}

// --- Platform API payload types ---

// Record61 is the JSON body for 6.1 (fault/predict write).
type Record61 struct {
	ID          string `json:"id"`
	MessageType string `json:"message_type"` // "0"=fault "1"=predict
	TrainType   string `json:"train_type"`
	TrainNo     string `json:"train_no"`
	Coach       string `json:"coach"`
	Location    string `json:"location"`
	Code        string `json:"code"`
	Station1    string `json:"station1"`
	Station2    string `json:"station2"`
	StartTime   string `json:"starttime"`
	EndTime     string `json:"endtime"`
	Subsystem   string `json:"subsystem"`
	LineName    string `json:"line_name"`
}

// Status66 is the JSON body for 6.6 (heartbeat).
type Status66 struct {
	MessageType string `json:"message_type"` // "500"
	Subsystem   string `json:"subsystem"`
	Status      string `json:"status"`   // "1"=ok "2"=partial "3"=degraded
	Remark      string `json:"remark"`
	Solution    string `json:"solution"`
	Time        string `json:"time"` // ms timestamp string
}

// LifeRecord67 is one entry in the 6.7 batch.
type LifeRecord67 struct {
	LineName     string `json:"lineName"`
	TrainType    string `json:"trainType"`
	TrainNo      string `json:"trainNo"`
	PartCode     string `json:"partCode"`
	ServiceTime  int64  `json:"serviceTime"`  // now ms
	ServiceValue int64  `json:"serviceValue"` // cumulative count or hours
	Mileage      int64  `json:"mileage"`      // always 0
	UseTime      int64  `json:"useTime"`      // start-of-service ms; 0 if unknown
	Flag         int    `json:"flag"`         // 2 = do not accumulate (send absolute value)
}

// rawInt safely extracts an integer from a map[string]any (JSON numbers are float64).
func rawInt(raw map[string]any, key string) int64 {
	v, ok := raw[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}
