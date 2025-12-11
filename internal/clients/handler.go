package clients

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	repo       *Repository
	adminToken string
}

func NewHandler(repo *Repository, adminToken string) *Handler {
	return &Handler{
		repo:       repo,
		adminToken: adminToken,
	}
}

// middleware simple para admin
func (h *Handler) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("X-Admin-Token")
		if token == "" || token != h.adminToken {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
		}
		return next(c)
	}
}

type createClientRequest struct {
	ClientUID string `json:"client_uid"` // opcional, si vacío generamos uno
	Provider  string `json:"provider"`   // "mercadopago", "stripe", "paypal"
}

type createClientResponse struct {
	ClientUID string `json:"client_uid"`
	Secret    string `json:"secret"` // lo mostramos solo una vez
	Provider  string `json:"provider"`
}

// POST /admin/clients
func (h *Handler) CreateClient(c echo.Context) error {
	var req createClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid body",
		})
	}

	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "provider is required",
		})
	}

	if req.ClientUID == "" {
		uid, err := generateRandomHex(8)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to generate client uid",
			})
		}
		req.ClientUID = uid
	}

	secret, err := generateRandomHex(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate secret",
		})
	}

	client, err := h.repo.Create(req.ClientUID, secret, req.Provider)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create client",
		})
	}

	return c.JSON(http.StatusCreated, createClientResponse{
		ClientUID: client.UID,
		Secret:    secret,
		Provider:  client.Provider,
	})
}

// GET /admin/clients
func (h *Handler) ListClients(c echo.Context) error {
	clients, err := h.repo.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to list clients",
		})
	}

	return c.JSON(http.StatusOK, clients)
}

func generateRandomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
