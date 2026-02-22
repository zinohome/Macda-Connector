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
}

func loadConfig() Config {
	return Config{
		KafkaBrokers:  splitCSV(getEnv("KAFKA_BROKERS", "redpanda-1:9092,redpanda-2:9092,redpanda-3:9092")),
		KafkaTopics:   splitCSV(getEnv("KAFKA_TOPICS", "signal-storage,signal-alarm,signal-predict,signal-life")),
		KafkaGroup:    getEnv("KAFKA_GROUP", "macda-storage-adapter-v1"),
		PGDSN:         getEnv("PG_DSN", "postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable"),
		BatchSize:     getEnvInt("BATCH_SIZE", 300),
		FlushInterval: time.Duration(getEnvInt("FLUSH_INTERVAL_MS", 300)) * time.Millisecond,
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
	}
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
