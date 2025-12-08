package clients

import (
	"context"
	"errors"

	"github.com/Kmicac/Webhook-Relay/internal/storage"
)

type Client struct {
	ID       int64
	UID      string
	Secret   string
	Provider string
}

type Repository struct {
	db *storage.PostgresStore
}

func NewRepository(store *storage.PostgresStore) *Repository {
	return &Repository{db: store}
}

func (r *Repository) FindByUID(uid string) (*Client, error) {
	row := r.db.DB.QueryRow(
		context.Background(),
		`SELECT id, client_uid, secret, provider
         FROM clients
         WHERE client_uid = $1`,
		uid,
	)

	var c Client
	err := row.Scan(&c.ID, &c.UID, &c.Secret, &c.Provider)
	if err != nil {
		return nil, errors.New("client not found")
	}

	return &c, nil
}
