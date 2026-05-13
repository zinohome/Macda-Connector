package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := loadConfig()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("[INFO] ground-reporter starting: platform=%s:%d subsystem=%s trainType=%s",
		cfg.PlatformIP, cfg.PlatformPort, cfg.SubsystemCode, cfg.TrainType)

	client := newPlatformClient(cfg)
	tracker := newAlarmTracker()
	lifeCache := newLifeCache()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// --- 6.1: signal-alarm ---
	wg.Add(1)
	go consumeTopic(ctx, &wg, cfg.KafkaBrokers,
		"signal-alarm", "ground-reporter-alarm",
		func(data []byte) {
			Handle61Alarm(ctx, client, tracker, cfg, data)
		},
	)

	// --- 6.1: signal-predict ---
	wg.Add(1)
	go consumeTopic(ctx, &wg, cfg.KafkaBrokers,
		"signal-predict", "ground-reporter-predict",
		func(data []byte) {
			Handle61Predict(ctx, client, tracker, cfg, data)
		},
	)

	// --- 6.7 per-action: signal-life ---
	wg.Add(1)
	go consumeTopic(ctx, &wg, cfg.KafkaBrokers,
		"signal-life", "ground-reporter-life",
		func(data []byte) {
			Handle67LifeAction(ctx, client, cfg, data)
		},
	)

	// --- 6.7 daily cache: signal-parsed ---
	wg.Add(1)
	go consumeTopic(ctx, &wg, cfg.KafkaBrokers,
		"signal-parsed", "ground-reporter-life-cache",
		func(data []byte) {
			var msg ParsedMsg
			if err := json.Unmarshal(data, &msg); err != nil {
				if cfg.LogLevel == "DEBUG" {
					log.Printf("[DEBUG] life-cache: bad json: %v", err)
				}
				return
			}
			lifeCache.Update(msg, cfg.TrainType)
		},
	)

	// --- 6.6: heartbeat timer ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		Run66Heartbeat(ctx, client, cfg)
	}()

	// --- 6.7: daily batch timer ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		Run67DailyBatch(ctx, client, lifeCache, cfg)
	}()

	// Graceful shutdown on SIGTERM / SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Printf("[INFO] ground-reporter shutting down...")
	cancel()
	wg.Wait()
	log.Printf("[INFO] ground-reporter stopped")
}
