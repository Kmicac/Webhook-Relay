package webhooks

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler es el controlador de webhooks de pagos.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// HandlePayment recibe y procesa webhooks de pago.
func (h *Handler) HandlePayment(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid body",
		})
	}

	provider := "mercadopago"

	ev := h.service.SavePaymentEvent(provider, string(body))

	// Responder rápido: esto es esencial para webhooks
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "received",
		"event_id":  ev.ID,
		"provider":  ev.Provider,
		"received":  ev.ReceivedAt,
	})
}

func (h *Handler) ListEvents(c echo.Context) error {
	events := h.service.ListEvents()
	return c.JSON(http.StatusOK, events)
}