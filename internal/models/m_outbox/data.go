package m_outbox

import (
	"cloud.google.com/go/spanner"
	"time"
)

// OutboxEvent represents the database model for outbox events
type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     string // JSON string
	Status      string
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// InsertMut creates a Spanner insert mutation for an outbox event
func (o *OutboxEvent) InsertMut() *spanner.Mutation {
	return spanner.Insert(
		TableName,
		[]string{
			EventID, EventType, AggregateID, Payload, Status, CreatedAt, ProcessedAt,
		},
		[]interface{}{
			o.EventID, o.EventType, o.AggregateID, o.Payload, o.Status, o.CreatedAt, o.ProcessedAt,
		},
	)
}

// UpdateMut creates a Spanner update mutation for an outbox event
// Note: columns must include EventID as the first column (primary key)
func (o *OutboxEvent) UpdateMut(columns []string) *spanner.Mutation {
	values := make([]interface{}, 0, len(columns))
	for _, col := range columns {
		switch col {
		case EventID:
			values = append(values, o.EventID)
		case EventType:
			values = append(values, o.EventType)
		case AggregateID:
			values = append(values, o.AggregateID)
		case Payload:
			values = append(values, o.Payload)
		case Status:
			values = append(values, o.Status)
		case ProcessedAt:
			values = append(values, o.ProcessedAt)
		}
	}

	return spanner.Update(
		TableName,
		columns,
		values,
	)
}

// DeleteMut creates a Spanner delete mutation for an outbox event
func (o *OutboxEvent) DeleteMut() *spanner.Mutation {
	return spanner.Delete(TableName, spanner.Key{o.EventID})
}

// TableName is the Spanner table name for outbox events
const TableName = "outbox_events"
