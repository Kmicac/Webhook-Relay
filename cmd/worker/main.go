package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kmicac/Webhook-Relay/internal/payments"
	"github.com/Kmicac/Webhook-Relay/internal/storage"
	"github.com/Kmicac/Webhook-Relay/internal/webhooks"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://webhookuser:webhookpass@localhost:5433/webhookrelay?sslmode=disable"
	}

	store := storage.NewPostgresStore(dsn)
	defer store.Close()

	webhookRepo := webhooks.NewRepository(store)
	paymentRepo := payments.NewRepository(store)
	paymentService := payments.NewService(paymentRepo)
	webhookService := webhooks.NewService(webhookRepo, paymentService)

	log.Println("[Worker] starting webhook processor loop")

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] shutting down")
			return
		default:
		}

		processed, err := webhookService.ProcessNextPending(ctx)
		if err != nil {
			log.Printf("[Worker] error processing event: %v\n", err)
		}

		if !processed {
			time.Sleep(2 * time.Second)
		}
	}
}
