package create_product

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"catalog-proj/internal/app/product/contracts"
	"catalog-proj/internal/app/product/domain"
	"catalog-proj/internal/models/m_outbox"
	"catalog-proj/internal/pkg/clock"

	"github.com/wuyiadepoju/commitplan"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
)

// Request represents the input for creating a product
type Request struct {
	Name        string
	Description string
	Category    string
	BasePrice   *domain.Money
}

// Response represents the output of creating a product
type Response struct {
	ProductID string
}

// Interactor handles the create product use case
type Interactor struct {
	repo      contracts.ProductRepository
	committer commitplan.Committer
	clock     clock.Clock
}

// NewInteractor creates a new create product interactor
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

// Execute creates a new product following the Golden Mutation Pattern
func (i *Interactor) Execute(ctx context.Context, req *Request) (*Response, error) {
	// Validate inputs
	if req.BasePrice == nil {
		return nil, fmt.Errorf("base_price is required")
	}
	// Validate price is positive (check sign of *big.Rat)
	priceRat := (*big.Rat)(*req.BasePrice)
	if priceRat.Sign() <= 0 {
		return nil, domain.ErrInvalidPrice
	}

	now := i.clock.Now()
	productID := uuid.New().String()

	// 1. Create aggregate (NewProduct sets initial status and emits ProductCreatedEvent)
	product := domain.NewProduct(
		productID,
		req.Name,
		req.Description,
		req.Category,
		req.BasePrice,
		now,
	)

	// 2. Get insert mutation from repo
	plan := commitplan.NewPlan()
	productMut := i.repo.InsertMut(product)
	plan.Add(productMut)

	// 3. Collect domain events → outbox mutations
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

	// 4. Apply plan via committer
	if err := i.committer.Apply(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// 5. Return product ID
	return &Response{
		ProductID: productID,
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
