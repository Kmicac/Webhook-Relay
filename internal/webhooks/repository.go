package webhooks

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/Kmicac/Webhook-Relay/internal/storage"
)

type Repository struct {
	db *storage.PostgresStore
}

func NewRepository(store *storage.PostgresStore) *Repository {
	return &Repository{db: store}
}

func (r *Repository) CreateEvent(ev *WebhookEvent) error {
	err := r.db.DB.QueryRow(
		context.Background(),
		`INSERT INTO webhook_events (provider, raw_body, processed, attempts)
         VALUES ($1, $2, FALSE, 0)
         RETURNING id, received_at`,
		ev.Provider, ev.RawBody,
	).Scan(&ev.ID, &ev.ReceivedAt)
	if err != nil {
		log.Printf("[WebhooksRepository] error creating webhook event: %v\n", err)
		return err
	}

	ev.Processed = false
	ev.ProcessedAt = nil
	ev.Attempts = 0
	ev.ErrorMessage = nil

	return nil
}

func (r *Repository) FetchNextPending(ctx context.Context) (*WebhookEvent, error) {
	tx, err := r.db.DB.Begin(ctx)
	if err != nil {
		log.Printf("[WebhooksRepository] error starting transaction: %v\n", err)
		return nil, err
	}

	var ev WebhookEvent
	err = tx.QueryRow(
		ctx,
		`SELECT id, provider, raw_body, received_at, processed, processed_at, attempts, error_message
         FROM webhook_events
         WHERE processed = FALSE
         ORDER BY id
         FOR UPDATE SKIP LOCKED
         LIMIT 1`,
	).Scan(
		&ev.ID,
		&ev.Provider,
		&ev.RawBody,
		&ev.ReceivedAt,
		&ev.Processed,
		&ev.ProcessedAt,
		&ev.Attempts,
		&ev.ErrorMessage,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return nil, nil
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		log.Printf("[WebhooksRepository] error fetching pending event: %v\n", err)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[WebhooksRepository] error committing pending fetch: %v\n", err)
		return nil, err
	}

	return &ev, nil
}

func (r *Repository) MarkProcessed(ctx context.Context, id int64) error {
	_, err := r.db.DB.Exec(
		ctx,
		`UPDATE webhook_events
         SET processed = TRUE,
             processed_at = NOW(),
             attempts = attempts + 1,
             error_message = NULL
         WHERE id = $1`,
		id,
	)
	if err != nil {
		log.Printf("[WebhooksRepository] error marking processed (id=%d): %v\n", id, err)
	}
	return err
}

func (r *Repository) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	_, err := r.db.DB.Exec(
		ctx,
		`UPDATE webhook_events
         SET processed = FALSE,
             processed_at = NULL,
             attempts = attempts + 1,
             error_message = SUBSTRING($2 FOR 500)
         WHERE id = $1`,
		id,
		errMsg,
	)
	if err != nil {
		log.Printf("[WebhooksRepository] error marking failed (id=%d): %v\n", id, err)
	}
	return err
}

// ListAll devuelve todos los eventos guardados en la tabla webhook_events.
func (r *Repository) ListAll() ([]WebhookEvent, error) {
	rows, err := r.db.DB.Query(
		context.Background(),
		`SELECT id, provider, raw_body, received_at, processed, processed_at, attempts, error_message
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
		if err := rows.Scan(
			&ev.ID,
			&ev.Provider,
			&ev.RawBody,
			&ev.ReceivedAt,
			&ev.Processed,
			&ev.ProcessedAt,
			&ev.Attempts,
			&ev.ErrorMessage,
		); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}

	return events, rows.Err()
}
