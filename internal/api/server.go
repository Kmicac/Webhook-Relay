package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Kmicac/webhookrelay/internal/storage"
	"github.com/Kmicac/webhookrelay/internal/webhooks"
)

func NewServer() *echo.Echo {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// DATABASE CONNECTION
	dsn := "postgres://webhookuser:webhookpass@localhost:5432/webhookrelay"
	store := storage.NewPostgresStore(dsn)

	// REPOSITORY 
	repo := webhooks.NewRepository(store)

	// SERVICE
	webhookService := webhooks.NewService(repo)

	// HANDLER 
	webhookHandler := webhooks.NewHandler(webhookService)

	// ROUTES
	e.POST("/webhooks/payments", webhookHandler.HandlePayment)
	e.GET("/webhooks/events", webhookHandler.ListEvents)

	return e
}
