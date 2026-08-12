package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokers  []string
	KafkaTopics   []string
	KafkaGroup    string
	PGDSN         string
	BatchSize     int
	FlushInterval time.Duration
	LogLevel      string
	// WriteFactEvent 控制是否把 event 消息写入 hvac.fact_event。
	// 默认 true（向后兼容 Plan B 全量替代场景）；在 lifecycle-only 部署里设为 false，
	// 以避免与 connect-event-writer（YAML Plan A）双重写入 fact_event。
	WriteFactEvent bool
	// WriteFactRaw 控制是否把 storage 消息写入 hvac.fact_raw。
	// 默认 true；lifecycle-only 部署通常只订阅 signal-predict，不会走此路径。
	WriteFactRaw bool
	// LifecycleSweeperInterval 是后台清扫器循环的间隔；<=0 关闭清扫器。
	// 参考 issue #26：现有 close 语义完全依赖 nb67_event_processor 的
	// prevPredictHadHits 状态机；生产车厢停发时 close 帧永远不到 → end_time NULL 永久滞留。
	// 清扫器兜底把停摆超过 LifecycleSweeperGrace 的活跃行 close 掉，end_time=last_seen_time。
	LifecycleSweeperInterval time.Duration
	// LifecycleSweeperGrace 是 last_seen_time 落后 wall clock 多久后被视为 "该 close" 的阈值。
	// 必须显著大于生产车厢正常帧间隔的 P99，避免误关正在活跃的预警。
	LifecycleSweeperGrace time.Duration
}

func loadConfig() Config {
	return Config{
		KafkaBrokers:   splitCSV(getEnv("KAFKA_BROKERS", "redpanda-1:9092,redpanda-2:9092,redpanda-3:9092")),
		KafkaTopics:    splitCSV(getEnv("KAFKA_TOPICS", "signal-storage,signal-alarm,signal-predict,signal-life")),
		KafkaGroup:     getEnv("KAFKA_GROUP", "macda-storage-adapter-v1"),
		PGDSN:          getEnv("PG_DSN", "postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable"),
		BatchSize:      getEnvInt("BATCH_SIZE", 300),
		FlushInterval:  time.Duration(getEnvInt("FLUSH_INTERVAL_MS", 300)) * time.Millisecond,
		LogLevel:       getEnv("LOG_LEVEL", "INFO"),
		WriteFactEvent:           getEnvBool("WRITE_FACT_EVENT", true),
		WriteFactRaw:             getEnvBool("WRITE_FACT_RAW", true),
		LifecycleSweeperInterval: time.Duration(getEnvInt("LIFECYCLE_SWEEPER_INTERVAL_SEC", 30)) * time.Second,
		LifecycleSweeperGrace:    time.Duration(getEnvInt("LIFECYCLE_SWEEPER_GRACE_SEC", 120)) * time.Second,
	}
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return number
}
