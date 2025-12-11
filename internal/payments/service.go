package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

type PaymentEvent struct {
	ExternalID   string     `json:"id"`
	Status       string     `json:"status"`
	StatusDetail string     `json:"status_detail"`
	Amount       float64    `json:"amount"`
	Currency     string     `json:"currency"`
	PayerEmail   string     `json:"payer_email"`
	ApprovedAt   *time.Time `json:"approved_at"`
	Provider     string     `json:"provider"`
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Process(rawBody []byte, webhookEventID int64, provider string) error {
	var payload map[string]interface{}

	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	var event PaymentEvent
	event.Provider = provider

	switch provider {
	case "mercadopago":
		event = parseMercadoPagoPayload(payload)
		event.Provider = provider

	case "stripe":
		event = parseStripePayload(payload)
		event.Provider = provider

	case "paypal":
		event = parsePaypalPayload(payload)
		event.Provider = provider

	default:
		event = PaymentEvent{
			ExternalID: getString(payload, "id"),
			Status:     getString(payload, "status"),
			Amount:     getFloat(payload, "amount"),
			Provider:   provider,
		}
	}

	log.Printf("[PaymentService] Parsed PaymentEvent (%s): %+v\n", provider, event)

	if err := s.repo.Save(event, webhookEventID); err != nil {
		return err
	}

	return nil
}

func parseMercadoPagoPayload(payload map[string]interface{}) PaymentEvent {
	ev := PaymentEvent{
		ExternalID:   getString(payload, "id"),
		Status:       getString(payload, "status"),
		StatusDetail: getString(payload, "status_detail"),
		Amount:       getFloat(payload, "transaction_amount"),
		Currency:     getString(payload, "currency_id"),
	}

	if payerRaw, ok := payload["payer"]; ok {
		if payerMap, ok := payerRaw.(map[string]interface{}); ok {
			ev.PayerEmail = getStringFromMap(payerMap, "email")
		}
	}

	dateApprovedStr := getString(payload, "date_approved")
	if dateApprovedStr != "" {
		if t, err := time.Parse(time.RFC3339, dateApprovedStr); err == nil {
			ev.ApprovedAt = &t
		}
	}

	if ev.Amount == 0 {
		ev.Amount = getFloat(payload, "amount")
	}

	return ev
}

func parseStripePayload(payload map[string]interface{}) PaymentEvent {
	ev := PaymentEvent{}

	dataObj, ok := payload["data"].(map[string]interface{})
	if !ok {
		ev.ExternalID = getString(payload, "id")
		ev.Status = getString(payload, "type")
		return ev
	}

	object, ok := dataObj["object"].(map[string]interface{})
	if !ok {
		ev.ExternalID = getString(payload, "id")
		ev.Status = getString(payload, "type")
		return ev
	}

	ev.ExternalID = getStringFromMap(object, "id")
	ev.Status = getStringFromMap(object, "status")

	// amount_received viene en centavos
	amountCents := getFloat(object, "amount_received")
	if amountCents != 0 {
		ev.Amount = amountCents / 100.0
	}

	ev.Currency = getStringFromMap(object, "currency")

	// payer email: charges.data[0].billing_details.email
	if chargesRaw, ok := object["charges"]; ok {
		if chargesMap, ok := chargesRaw.(map[string]interface{}); ok {
			if dataSliceRaw, ok := chargesMap["data"]; ok {
				if dataSlice, ok := dataSliceRaw.([]interface{}); ok && len(dataSlice) > 0 {
					if firstCharge, ok := dataSlice[0].(map[string]interface{}); ok {
						if billingRaw, ok := firstCharge["billing_details"]; ok {
							if billing, ok := billingRaw.(map[string]interface{}); ok {
								ev.PayerEmail = getStringFromMap(billing, "email")
							}
						}
					}
				}
			}
		}
	}

	return ev
}

func parsePaypalPayload(payload map[string]interface{}) PaymentEvent {
	ev := PaymentEvent{}

	ev.ExternalID = getString(payload, "id")
	ev.Status = getString(payload, "status")
	ev.StatusDetail = getString(payload, "status_detail")

	// amount.value + amount.currency_code
	if amountRaw, ok := payload["amount"]; ok {
		if amountMap, ok := amountRaw.(map[string]interface{}); ok {
			// value viene como string
			valStr := getStringFromMap(amountMap, "value")
			if valStr != "" {
				if f, err := strconv.ParseFloat(valStr, 64); err == nil {
					ev.Amount = f
				}
			}
			ev.Currency = getStringFromMap(amountMap, "currency_code")
		}
	}

	if payerRaw, ok := payload["payer"]; ok {
		if payerMap, ok := payerRaw.(map[string]interface{}); ok {
			ev.PayerEmail = getStringFromMap(payerMap, "email_address")
		}
	}

	updateTime := getString(payload, "update_time")
	if updateTime != "" {
		if t, err := time.Parse(time.RFC3339, updateTime); err == nil {
			ev.ApprovedAt = &t
		}
	}

	return ev
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case int:
			return float64(t)
		}
	}
	return 0
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
