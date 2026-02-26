package activate_product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"catalog-proj/internal/app/product/contracts"
	"catalog-proj/internal/app/product/domain"
	"catalog-proj/internal/models/m_outbox"
	"catalog-proj/internal/pkg/clock"
	"github.com/wuyiadepoju/commitplan"
	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
)

// Request represents the input for activating a product
type Request struct {
	ProductID string
}

// Response represents the output of activating a product
type Response struct {
	ProductID string
}

// Interactor handles the activate product use case
type Interactor struct {
	repo      contracts.ProductRepository
	committer commitplan.Committer
	clock     clock.Clock
}

// NewInteractor creates a new activate product interactor
func NewInteractor(
	repo contracts.ProductRepository,
	committer commitplan.Committer,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		repo:      repo,
		committer: committer,
		clock:     clock,
	}
}

// Execute activates a product following the Golden Mutation Pattern
func (i *Interactor) Execute(ctx context.Context, req *Request) (*Response, error) {
	// 1. Load aggregate
	product, err := i.repo.Load(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to load product: %w", err)
	}

	// 2. Call domain method
	now := i.clock.Now()
	if err := product.Activate(now); err != nil {
		return nil, fmt.Errorf("failed to activate product: %w", err)
	}

	// 3. Get update mutation
	plan := commitplan.NewPlan()
	productMut := i.repo.UpdateMut(product)
	if productMut != nil {
		plan.Add(productMut)
	}

	// 4. Collect events → outbox
	events := product.DomainEvents()
	for _, event := range events {
		outboxMut, err := i.eventToOutboxMutation(event, now)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox event: %w", err)
		}
		if outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// 5. Apply plan
	if len(plan.Mutations()) > 0 {
		if err := i.committer.Apply(ctx, plan); err != nil {
			return nil, fmt.Errorf("failed to activate product: %w", err)
		}
	}

	// 6. Return product ID
	return &Response{
		ProductID: req.ProductID,
	}, nil
}

// eventToOutboxMutation converts a domain event to an outbox mutation
func (i *Interactor) eventToOutboxMutation(event domain.DomainEvent, now time.Time) (*spanner.Mutation, error) {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data for event %s: %w", event.EventName(), err)
	}

	// Extract aggregate_id from event data (product_id)
	aggregateID := ""
	if data := event.EventData(); data != nil {
		if pid, ok := data["product_id"].(string); ok {
			aggregateID = pid
		}
	}

	outboxEvent := &m_outbox.OutboxEvent{
		EventID:     uuid.New().String(),
		EventType:   event.EventName(),
		AggregateID: aggregateID,
		Payload:     string(eventData),
		Status:      "pending",
		CreatedAt:   now,
		ProcessedAt: nil,
	}

	return outboxEvent.InsertMut(), nil
}
