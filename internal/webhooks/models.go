package webhooks

import "time"

type WebhookEvent struct {
	ID           int64      `db:"id" json:"id"`
	Provider     string     `db:"provider" json:"provider"`
	RawBody      string     `db:"raw_body" json:"raw_body"`
	ReceivedAt   time.Time  `db:"received_at" json:"received_at"`
	Processed    bool       `db:"processed" json:"processed"`
	ProcessedAt  *time.Time `db:"processed_at" json:"processed_at,omitempty"`
	Attempts     int        `db:"attempts" json:"attempts"`
	ErrorMessage *string    `db:"error_message" json:"error_message,omitempty"`
}
