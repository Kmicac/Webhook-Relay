package webhooks

import (
	"context"
	"log"

	"github.com/Kmicac/Webhook-Relay/internal/payments"
)

type Service struct {
	repo           *Repository
	paymentService *payments.Service
}

func NewService(repo *Repository, paymentService *payments.Service) *Service {
	return &Service{
		repo:           repo,
		paymentService: paymentService,
	}
}

func (s *Service) EnqueueEvent(provider string, rawBody string) (*WebhookEvent, error) {
	ev := &WebhookEvent{
		Provider:  provider,
		RawBody:   rawBody,
		Processed: false,
		Attempts:  0,
	}

	if err := s.repo.CreateEvent(ev); err != nil {
		log.Printf("[WebhookService] error enqueuing webhook event: %v\n", err)
		return nil, err
	}

	return ev, nil
}

func (s *Service) ProcessNextPending(ctx context.Context) (bool, error) {
	ev, err := s.repo.FetchNextPending(ctx)
	if err != nil {
		return false, err
	}

	if ev == nil {
		return false, nil
	}

	log.Printf("[Worker] processing event id=%d provider=%s\n", ev.ID, ev.Provider)

	if err := s.paymentService.Process([]byte(ev.RawBody), ev.ID, ev.Provider); err != nil {
		log.Printf("[WebhookService] error processing event id=%d: %v\n", ev.ID, err)
		_ = s.repo.MarkFailed(ctx, ev.ID, err.Error())
		return true, err
	}

	if err := s.repo.MarkProcessed(ctx, ev.ID); err != nil {
		return true, err
	}

	log.Printf("[Worker] processed event id=%d\n", ev.ID)

	return true, nil
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
