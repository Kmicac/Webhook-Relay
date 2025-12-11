package clients

import (
	"context"
	"errors"
	"log"

	"github.com/Kmicac/Webhook-Relay/internal/storage"
)

type Client struct {
	ID       int64  `json:"id"`
	UID      string `json:"client_uid"`
	Secret   string `json:"-"`
	Provider string `json:"provider"`
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
	if err := row.Scan(&c.ID, &c.UID, &c.Secret, &c.Provider); err != nil {
		return nil, errors.New("client not found")
	}

	return &c, nil
}

func (r *Repository) List() ([]Client, error) {
	rows, err := r.db.DB.Query(
		context.Background(),
		`SELECT id, client_uid, provider
         FROM clients
         ORDER BY id DESC`,
	)
	if err != nil {
		log.Printf("[ClientsRepository] List ERROR: %v", err)
		return nil, err
	}
	defer rows.Close()

	var result []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.UID, &c.Provider); err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, nil
}

func (r *Repository) Create(uid, secret, provider string) (*Client, error) {
	row := r.db.DB.QueryRow(
		context.Background(),
		`INSERT INTO clients (client_uid, secret, provider)
         VALUES ($1, $2, $3)
         RETURNING id, client_uid, secret, provider`,
		uid, secret, provider,
	)

	var c Client
	if err := row.Scan(&c.ID, &c.UID, &c.Secret, &c.Provider); err != nil {
		return nil, err
	}

	return &c, nil
}
