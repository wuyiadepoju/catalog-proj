package domain

import "time"

type DomainEvent interface {
	EventName() string
	EventData() map[string]interface{}
}

type ProductCreatedEvent struct {
	ProductID string
	Name      string
	Category  string
	BasePrice *Money
	CreatedAt time.Time
}

func (e *ProductCreatedEvent) EventName() string {
	return "product_created"
}

func (e *ProductCreatedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id": e.ProductID,
		"name":       e.Name,
		"category":   e.Category,
		"created_at": e.CreatedAt,
	}
}

type ProductUpdatedEvent struct {
	ProductID     string
	UpdatedAt     time.Time
	ChangedFields []string
}

func (e *ProductUpdatedEvent) EventName() string {
	return "product_updated"
}

func (e *ProductUpdatedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id":     e.ProductID,
		"updated_at":     e.UpdatedAt,
		"changed_fields": e.ChangedFields,
	}
}

type DiscountAppliedEvent struct {
	ProductID  string
	DiscountID string
	AppliedAt  time.Time
}

func (e *DiscountAppliedEvent) EventName() string {
	return "discount_applied"
}

func (e *DiscountAppliedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id":  e.ProductID,
		"discount_id": e.DiscountID,
		"applied_at":  e.AppliedAt,
	}
}

type ProductActivatedEvent struct {
	ProductID   string
	ActivatedAt time.Time
}

func (e *ProductActivatedEvent) EventName() string {
	return "product_activated"
}

func (e *ProductActivatedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id":   e.ProductID,
		"activated_at": e.ActivatedAt,
	}
}

type ProductDeactivatedEvent struct {
	ProductID     string
	DeactivatedAt time.Time
}

func (e *ProductDeactivatedEvent) EventName() string {
	return "product_deactivated"
}

func (e *ProductDeactivatedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id":     e.ProductID,
		"deactivated_at": e.DeactivatedAt,
	}
}

type ProductArchivedEvent struct {
	ProductID  string
	ArchivedAt time.Time
}

func (e *ProductArchivedEvent) EventName() string {
	return "product_archived"
}

func (e *ProductArchivedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id":  e.ProductID,
		"archived_at": e.ArchivedAt,
	}
}

type DiscountRemovedEvent struct {
	ProductID string
	RemovedAt time.Time
}

func (e *DiscountRemovedEvent) EventName() string {
	return "discount_removed"
}

func (e *DiscountRemovedEvent) EventData() map[string]interface{} {
	return map[string]interface{}{
		"product_id": e.ProductID,
		"removed_at": e.RemovedAt,
	}
}
