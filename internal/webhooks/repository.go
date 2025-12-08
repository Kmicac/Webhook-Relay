package webhooks

import (
	"context"

	"github.com/Kmicac/Webhook-Relay/internal/storage"
)

type Repository struct {
	db *storage.PostgresStore
}

func NewRepository(store *storage.PostgresStore) *Repository {
	return &Repository{db: store}
}

func (r *Repository) Save(provider string, rawBody string) (int64, error) {
	var id int64

	err := r.db.DB.QueryRow(
		context.Background(),
		`INSERT INTO webhook_events (provider, raw_body) 
         VALUES ($1, $2) 
         RETURNING id`,
		provider, rawBody,
	).Scan(&id)

	return id, err
}

// ListAll devuelve todos los eventos guardados en la tabla webhook_events.
func (r *Repository) ListAll() ([]WebhookEvent, error) {
	rows, err := r.db.DB.Query(
		context.Background(),
		`SELECT id, provider, raw_body, received_at
         FROM webhook_events
         ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []WebhookEvent

	for rows.Next() {
		var ev WebhookEvent
		if err := rows.Scan(&ev.ID, &ev.Provider, &ev.RawBody, &ev.ReceivedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}

	return events, rows.Err()
}
