package main

import (
	"context"
	"log"
	"strconv"
	"time"
)

const path66 = "/gate/METRO-SELFCHECK-SUBSYSTEM/api/faultRecordsSubsystem/saveStatus"

// Run66Heartbeat sends a heartbeat to the platform every HeartbeatIntervalMin minutes.
// Runs until ctx is cancelled.
func Run66Heartbeat(ctx context.Context, client *PlatformClient, cfg Config) {
	interval := time.Duration(cfg.HeartbeatIntervalMin) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send one immediately on startup so the platform sees us right away.
	sendHeartbeat(ctx, client, cfg)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendHeartbeat(ctx, client, cfg)
		}
	}
}

func sendHeartbeat(ctx context.Context, client *PlatformClient, cfg Config) {
	payload := []Status66{
		{
			MessageType: "500",
			Subsystem:   cfg.SubsystemCode,
			Status:      "1",
			Remark:      "",
			Solution:    "",
			Time:        strconv.FormatInt(nowMs(), 10),
		},
	}
	if err := client.PostJSON(ctx, path66, payload); err != nil {
		log.Printf("[ERROR] 6.6 heartbeat POST failed: %v", err)
		return
	}
	log.Printf("[INFO] 6.6 heartbeat sent")
}
