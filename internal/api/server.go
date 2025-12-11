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
	dsn := "postgres://webhookuser:webhookpass@localhost:5433/webhookrelay?sslmode=disable"
	store := storage.NewPostgresStore(dsn)

	// REPOSITORY
	repo := webhooks.NewRepository(store)

	// SERVICE
	clientRepo := clients.NewRepository(store)
	paymentRepo := payments.NewRepository(store)
	paymentService := payments.NewService(paymentRepo)
	webhookService := webhooks.NewService(repo, paymentService)

	// admin token
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = "dev-admin-token"
	}

	// HANDLER
	webhookHandler := webhooks.NewHandler(webhookService, clientRepo)
	clientHandler := clients.NewHandler(clientRepo, adminToken)

	// ROUTES
	e.POST("/webhooks/:client_id/:provider/payments", webhookHandler.HandlePayment)
	e.GET("/webhooks/events", webhookHandler.ListEvents)

	// ADMIN CLIENTS
	adminGroup := e.Group("/admin", clientHandler.RequireAdmin)
	adminGroup.POST("/clients", clientHandler.CreateClient)
	adminGroup.GET("/clients", clientHandler.ListClients)

	return e
}
