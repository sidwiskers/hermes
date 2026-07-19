package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/dedupe"
	"github.com/sidwiskers/hermes/observe"
	"github.com/sidwiskers/hermes/ratelimit"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	metrics := new(observe.Metrics)
	bot := hermes.New(
		os.Getenv("BOT_TOKEN"),
		hermes.WithMaxConcurrentUpdates(64),
		hermes.WithAPIObserver(metrics),
	)
	claims := dedupe.NewMemory(dedupe.MemoryConfig{MaxEntries: 1_000_000, Shards: 64})
	deduplicator := dedupe.New(claims, dedupe.WithNamespace("primary"), dedupe.WithTTL(24*time.Hour))
	limiter, err := ratelimit.New(10, time.Second,
		ratelimit.WithBurst(20),
		ratelimit.WithMaxKeys(100_000),
	)
	if err != nil {
		log.Fatal(err)
	}

	bot.Use(
		hermes.Recover(),
		observe.Middleware(metrics),
		deduplicator.Middleware(),
		limiter.Middleware(),
	)
	bot.Command("start", func(c *hermes.Context) error { return c.Send("Ready.") })

	go sweep(ctx, claims, limiter)
	if err := bot.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
	log.Printf("final metrics: %+v", metrics.Snapshot())
}

func sweep(ctx context.Context, claims *dedupe.Memory, limiter *ratelimit.Limiter) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			claims.Sweep()
			limiter.Sweep()
		}
	}
}
