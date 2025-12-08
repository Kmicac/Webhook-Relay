package webhooks

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Kmicac/Webhook-Relay/internal/clients"
	"github.com/Kmicac/Webhook-Relay/internal/payments"
)

// Handler es el controlador de webhooks de pagos.
type Handler struct {
	service        *Service
	secret         []byte
	paymentService *payments.Service
	clientRepo     *clients.Repository
}

func NewHandler(service *Service, secret string, paymentService *payments.Service, clientRepo *clients.Repository) *Handler {
	return &Handler{
		service:        service,
		secret:         []byte(secret),
		paymentService: paymentService,
		clientRepo:     clientRepo,
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

	clientID := c.Param("client_id")
	provider := c.Param("provider")

	client, err := h.clientRepo.FindByUID(clientID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid client",
		})
	}

	signature := c.Request().Header.Get("X-Signature")

	switch provider {
	case "mercadopago":
		if !VerifyMPSignature([]byte(client.Secret), signature, body) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid signature",
			})
		}
	case "stripe":
		// TODO: implementar verificación real de Stripe más adelante
		// Por ahora podrías aceptar sin firma o loguear:
		// log.Println("[WARN] Stripe signature verification not implemented yet")
	case "paypal":
		// TODO: implementar verificación real de PayPal más adelante
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "unsupported provider",
		})
	}

	ev := h.service.SavePaymentEvent(provider, string(body))

	h.paymentService.Process(body, ev.ID, provider)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "received",
		"event_id": ev.ID,
		"provider": ev.Provider,
		"received": ev.ReceivedAt,
	})
}

func (h *Handler) ListEvents(c echo.Context) error {
	events := h.service.ListEvents()
	return c.JSON(http.StatusOK, events)
}
