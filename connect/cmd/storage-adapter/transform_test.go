package main

import (
	"testing"
	"time"
)

// GitHub #24 / RET-46 回归：裸时间戳 "2026-05-28 16:28:46"（实为 Asia/Shanghai 本地时间）
// 必须按 +08:00 解析后归一化为 UTC，而非默认 UTC 解读 → 比真实时刻多 8h。
func TestParseTimeText_NakedTimestampIsShanghaiLocal(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		wantHM string // UTC 小时:分钟，期望比本地 -08:00
	}{
		{"space layout", "2026-05-28 16:28:46", "08:28"},
		{"T layout", "2026-05-28T16:28:46", "08:28"},
		{"loose layout", "2026-5-28 16:28:46", "08:28"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseTimeText(c.input)
			if !ok {
				t.Fatalf("parseTimeText(%q) returned ok=false", c.input)
			}
			if got.Location() != time.UTC {
				t.Errorf("parseTimeText(%q).Location() = %v, want UTC", c.input, got.Location())
			}
			gotHM := got.Format("15:04")
			if gotHM != c.wantHM {
				t.Errorf("parseTimeText(%q) = %s (UTC), want %s (UTC)", c.input, gotHM, c.wantHM)
			}
		})
	}
}

// 带时区的时间戳必须保留原本时区信息，不能被二次本地化。
func TestParseTimeText_TimezoneAwareNotDoubleShifted(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		wantHM string // UTC 期望值
	}{
		{"RFC3339 +08:00", "2026-05-28T16:28:46+08:00", "08:28"},
		{"RFC3339 Z", "2026-05-28T16:28:46Z", "16:28"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseTimeText(c.input)
			if !ok {
				t.Fatalf("parseTimeText(%q) returned ok=false", c.input)
			}
			gotHM := got.Format("15:04")
			if gotHM != c.wantHM {
				t.Errorf("parseTimeText(%q) = %s (UTC), want %s (UTC)", c.input, gotHM, c.wantHM)
			}
		})
	}
}
