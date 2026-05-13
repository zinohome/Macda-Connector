package main

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Platform HTTP endpoint
	PlatformIP             string
	PlatformPort           int
	PlatformApiKey         string // X-Api-Key header; empty = no auth
	PlatformTimeoutSec     int
	PlatformRetryMax       int
	PlatformRetryBackoffMs int

	// HVAC subsystem identity
	SubsystemCode string // fixed "5"
	TrainType     string // e.g. "B"

	// Heartbeat / daily batch schedule
	HeartbeatIntervalMin int
	DailyBatchHour       int // 0-23, default 0 (midnight)

	// Kafka
	KafkaBrokers []string

	LogLevel string
}

func loadConfig() Config {
	return Config{
		PlatformIP:             getEnv("PLATFORM_IP", "10.12.48.187"),
		PlatformPort:           getEnvInt("PLATFORM_PORT", 8188),
		PlatformApiKey:         getEnv("PLATFORM_API_KEY", ""),
		PlatformTimeoutSec:     getEnvInt("PLATFORM_TIMEOUT_SEC", 10),
		PlatformRetryMax:       getEnvInt("PLATFORM_RETRY_MAX", 3),
		PlatformRetryBackoffMs: getEnvInt("PLATFORM_RETRY_BACKOFF_MS", 1000),

		SubsystemCode: getEnv("SUBSYSTEM_CODE", "5"),
		TrainType:     getEnv("TRAIN_TYPE", "B"),

		HeartbeatIntervalMin: getEnvInt("HEARTBEAT_INTERVAL_MIN", 10),
		DailyBatchHour:       getEnvInt("DAILY_BATCH_HOUR", 0),

		KafkaBrokers: splitCSV(getEnv("KAFKA_BROKERS", "redpanda-1:9092,redpanda-2:9092,redpanda-3:9092")),

		LogLevel: getEnv("LOG_LEVEL", "INFO"),
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
