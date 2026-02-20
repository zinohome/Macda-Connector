package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := loadConfig()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.PGDSN)
	if err != nil {
		log.Fatalf("failed to connect timescaledb: %v", err)
	}
	defer pool.Close()

	service := &adapter{
		cfg:  cfg,
		pool: pool,
	}
	if err := service.run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("storage-adapter stopped with error: %v", err)
	}
}
