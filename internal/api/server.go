package api

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/Kmicac/Webhook-Relay/internal/clients"
	"github.com/Kmicac/Webhook-Relay/internal/payments"
	"github.com/Kmicac/Webhook-Relay/internal/storage"
	"github.com/Kmicac/Webhook-Relay/internal/webhooks"
)

func NewServer() *echo.Echo {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// DATABASE CONNECTION
	dsn := "postgres://webhookuser:webhookpass@localhost:5433/Webhook-Relay"
	store := storage.NewPostgresStore(dsn)

	// REPOSITORY
	repo := webhooks.NewRepository(store)

	// SERVICE
	webhookService := webhooks.NewService(repo)
	paymentRepo := payments.NewRepository(store)
	paymentService := payments.NewService(paymentRepo)
	clientRepo := clients.NewRepository(store)

	// SECRET PARA WEBHOOKS
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		secret = "my-hiper-secret-key"
	}

	// HANDLER
	webhookHandler := webhooks.NewHandler(webhookService, secret, paymentService, clientRepo)

	// ROUTES
	e.POST("/webhooks/:client_id/:provider/payments", webhookHandler.HandlePayment)
	e.GET("/webhooks/events", webhookHandler.ListEvents)

	return e
}
