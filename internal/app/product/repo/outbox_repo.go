package repo

import (
	"catalog-proj/internal/models/m_outbox"
	"cloud.google.com/go/spanner"
)

// OutboxRepository handles outbox event mutations
type OutboxRepository struct{}

// NewOutboxRepository creates a new outbox repository
func NewOutboxRepository() *OutboxRepository {
	return &OutboxRepository{}
}

// InsertMut creates a Spanner insert mutation for an outbox event
func (r *OutboxRepository) InsertMut(event *m_outbox.OutboxEvent) *spanner.Mutation {
	return event.InsertMut()
}
