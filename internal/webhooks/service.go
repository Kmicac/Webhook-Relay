package webhooks

import (
	"log"
	"time"
)

type Service struct {
	repo *Repository
}

type WebhookEvent struct {
	ID        int64     `json:"id"`
	Provider  string    `json:"provider"`
	RawBody   string    `json:"raw_body"`
	ReceivedAt time.Time `json:"received_at"`
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SavePaymentEvent(provider string, rawBody string) WebhookEvent {
	id, err := s.repo.Save(provider, rawBody)
	if err != nil {
		log.Printf("[WebhookService] Error guardando webhook: %v\n", err)
	}

	return WebhookEvent{
		ID:        id,
		Provider:  provider,
		RawBody:   rawBody,
		ReceivedAt: time.Now(),
	}
}

// ListEvents returns the events saved from the repository.
func (s *Service) ListEvents() []WebhookEvent {
	events, err := s.repo.ListAll()
	if err != nil {
		log.Printf("[WebhookService] Error listando eventos: %v\n", err)
		return []WebhookEvent{}
	}
	return events
}
