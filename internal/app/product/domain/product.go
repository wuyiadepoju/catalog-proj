package domain

import "time"

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

const (
	FieldDiscount    = "discount"
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldStatus      = "status"
	FieldArchivedAt  = "archived_at"
)

type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	changes     ChangeTracker
	events      []DomainEvent
	archivedAt  *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func NewProduct(id, name, description, category string, basePrice *Money) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		changes:     ChangeTracker{dirtyFields: make(map[string]bool)},
		events:      []DomainEvent{},
	}
}

// Business method (pure logic)
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {

	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}

	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}

	// Enforce business rule: Only one active discount per product at a time
	if p.discount != nil && p.discount.IsValidAt(now) {
		return ErrDiscountAlreadyActive
	}

	p.discount = discount
	p.changes.MarkDirty(FieldDiscount)
	p.events = append(p.events, &DiscountAppliedEvent{
		ProductID:  p.id,
		DiscountID: discount.ID,
		AppliedAt:  now,
	})
	return nil
}

// Getters (encapsulation)
func (p *Product) ID() string {
	return p.id
}

func (p *Product) Name() string {
	return p.name
}

func (p *Product) Description() string {
	return p.description
}

func (p *Product) Category() string {
	return p.category
}

func (p *Product) BasePrice() *Money {
	return p.basePrice
}

func (p *Product) Discount() *Discount {
	return p.discount
}

func (p *Product) Status() ProductStatus {
	return p.status
}

func (p *Product) Changes() ChangeTracker {
	return p.changes
}

func (p *Product) DomainEvents() []DomainEvent {
	return p.events
}

func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Product) ArchivedAt() *time.Time {
	return p.archivedAt
}

// ReconstructProduct creates a Product from persisted data
// This is used by the repository layer to reconstruct domain objects from the database
func ReconstructProduct(
	id string,
	name string,
	description string,
	category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	archivedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		changes:     ChangeTracker{dirtyFields: make(map[string]bool)},
		events:      []DomainEvent{},
		archivedAt:  archivedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// UpdateDetails updates the product's name, description, and category
func (p *Product) UpdateDetails(name, description, category string) error {
	if p.archivedAt != nil {
		return ErrProductAlreadyArchived
	}

	if name != p.name {
		p.name = name
		p.changes.MarkDirty(FieldName)
	}
	if description != p.description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
	}
	if category != p.category {
		p.category = category
		p.changes.MarkDirty(FieldCategory)
	}

	return nil
}

// Activate activates the product
func (p *Product) Activate() error {
	if p.archivedAt != nil {
		return ErrProductAlreadyArchived
	}

	if p.status == ProductStatusActive {
		return nil // Already active
	}

	p.status = ProductStatusActive
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, &ProductActivatedEvent{
		ProductID:   p.id,
		ActivatedAt: time.Now(),
	})

	return nil
}

// Deactivate deactivates the product
func (p *Product) Deactivate() error {
	if p.archivedAt != nil {
		return ErrProductAlreadyArchived
	}

	// Check if product has an active discount
	if p.discount != nil {
		now := time.Now()
		if p.discount.IsValidAt(now) {
			return ErrProductHasActiveDiscount
		}
	}

	if p.status == ProductStatusInactive {
		return nil // Already inactive
	}

	p.status = ProductStatusInactive
	p.changes.MarkDirty(FieldStatus)
	p.events = append(p.events, &ProductDeactivatedEvent{
		ProductID:     p.id,
		DeactivatedAt: time.Now(),
	})

	return nil
}

// Archive archives the product
func (p *Product) Archive(now time.Time) error {
	if p.archivedAt != nil {
		return ErrProductAlreadyArchived
	}

	p.archivedAt = &now
	p.changes.MarkDirty(FieldArchivedAt)
	p.events = append(p.events, &ProductArchivedEvent{
		ProductID:  p.id,
		ArchivedAt: now,
	})

	return nil
}

// RemoveDiscount removes the discount from the product
func (p *Product) RemoveDiscount() error {
	if p.discount == nil {
		return nil // No discount to remove
	}

	p.discount = nil
	p.changes.MarkDirty(FieldDiscount)
	p.events = append(p.events, &DiscountRemovedEvent{
		ProductID: p.id,
		RemovedAt: time.Now(),
	})

	return nil
}

type ChangeTracker struct {
	dirtyFields map[string]bool
}

func (ct *ChangeTracker) MarkDirty(field string) {
	if ct.dirtyFields == nil {
		ct.dirtyFields = make(map[string]bool)
	}
	ct.dirtyFields[field] = true
}

func (ct *ChangeTracker) Dirty(field string) bool {
	if ct.dirtyFields == nil {
		return false
	}
	return ct.dirtyFields[field]
}
