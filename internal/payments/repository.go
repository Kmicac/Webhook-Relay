package payments

import (
	"context"
	"log"

	"github.com/Kmicac/Webhook-Relay/internal/storage"
)

type Repository struct {
	db *storage.PostgresStore
}

func NewRepository(store *storage.PostgresStore) *Repository {
	return &Repository{db: store}
}

func (r *Repository) Save(event PaymentEvent, webhookEventID int64) error {
	_, err := r.db.DB.Exec(
		context.Background(),
		`INSERT INTO payments (
            external_id,
            status,
            status_detail,
            amount,
            currency,
            payer_email,
            approved_at,
            provider,
            webhook_event_id
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		event.ExternalID,
		event.Status,
		event.StatusDetail,
		event.Amount,
		event.Currency,
		event.PayerEmail,
		event.ApprovedAt,
		event.Provider,
		webhookEventID,
	)

	if err != nil {
		log.Printf("[PaymentRepository] error saving payment: %v", err)
	}

	return err
}
