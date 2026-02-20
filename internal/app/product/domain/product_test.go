package domain

import (
	"testing"
	"time"
)

func TestNewProduct(t *testing.T) {
	id := "product-1"
	name := "Test Product"
	description := "Test Description"
	category := "Electronics"
	basePrice := NewMoney(10000, "USD") // $100.00

	product := NewProduct(id, name, description, category, basePrice)

	if product.ID() != id {
		t.Errorf("ID() = %s, want %s", product.ID(), id)
	}
	if product.Name() != name {
		t.Errorf("Name() = %s, want %s", product.Name(), name)
	}
	if product.Description() != description {
		t.Errorf("Description() = %s, want %s", product.Description(), description)
	}
	if product.Category() != category {
		t.Errorf("Category() = %s, want %s", product.Category(), category)
	}
	if product.BasePrice() != basePrice {
		t.Errorf("BasePrice() = %v, want %v", product.BasePrice(), basePrice)
	}
	if product.Discount() != nil {
		t.Error("Discount() should be nil for new product")
	}
	if product.Status() != "" {
		t.Errorf("Status() = %s, want empty string", product.Status())
	}
	if len(product.DomainEvents()) != 0 {
		t.Errorf("DomainEvents() should be empty, got %d events", len(product.DomainEvents()))
	}
}

func TestProduct_ApplyDiscount(t *testing.T) {
	basePrice := NewMoney(10000, "USD") // $100.00
	product := NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
	
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	discount := &Discount{
		ID:        "discount-1",
		Amount:    NewMoney(10, 100), // 10%
		StartDate: startDate,
		EndDate:   endDate,
	}

	t.Run("successfully applies discount to active product", func(t *testing.T) {
		product.status = ProductStatusActive
		err := product.ApplyDiscount(discount, now)
		if err != nil {
			t.Errorf("ApplyDiscount() error = %v, want nil", err)
		}
		if product.Discount() == nil {
			t.Error("Discount() should not be nil after applying")
		}
		if product.Discount().ID != discount.ID {
			t.Errorf("Discount().ID = %s, want %s", product.Discount().ID, discount.ID)
		}
		if !product.Changes().Dirty(FieldDiscount) {
			t.Error("FieldDiscount should be marked as dirty")
		}
		if len(product.DomainEvents()) != 1 {
			t.Errorf("DomainEvents() length = %d, want 1", len(product.DomainEvents()))
		}
		event := product.DomainEvents()[0]
		if event.EventName() != "discount_applied" {
			t.Errorf("EventName() = %s, want discount_applied", event.EventName())
		}
	})

	t.Run("fails to apply discount to inactive product", func(t *testing.T) {
		product2 := NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		product2.status = ProductStatusInactive
		err := product2.ApplyDiscount(discount, now)
		if err != ErrProductNotActive {
			t.Errorf("ApplyDiscount() error = %v, want ErrProductNotActive", err)
		}
		if product2.Discount() != nil {
			t.Error("Discount() should be nil when application fails")
		}
	})

	t.Run("fails to apply discount with invalid period", func(t *testing.T) {
		product3 := NewProduct("product-3", "Test", "Desc", "Cat", basePrice)
		product3.status = ProductStatusActive
		invalidDiscount := &Discount{
			ID:        "discount-2",
			Amount:    NewMoney(10, 100),
			StartDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2024, 2, 28, 23, 59, 59, 0, time.UTC),
		}
		err := product3.ApplyDiscount(invalidDiscount, now)
		if err != ErrInvalidDiscountPeriod {
			t.Errorf("ApplyDiscount() error = %v, want ErrInvalidDiscountPeriod", err)
		}
	})
}

func TestProduct_UpdateDetails(t *testing.T) {
	basePrice := NewMoney(10000, "USD")
	product := NewProduct("product-1", "Old Name", "Old Desc", "Old Cat", basePrice)

	t.Run("successfully updates details", func(t *testing.T) {
		err := product.UpdateDetails("New Name", "New Desc", "New Cat")
		if err != nil {
			t.Errorf("UpdateDetails() error = %v, want nil", err)
		}
		if product.Name() != "New Name" {
			t.Errorf("Name() = %s, want New Name", product.Name())
		}
		if product.Description() != "New Desc" {
			t.Errorf("Description() = %s, want New Desc", product.Description())
		}
		if product.Category() != "New Cat" {
			t.Errorf("Category() = %s, want New Cat", product.Category())
		}
		if !product.Changes().Dirty(FieldName) {
			t.Error("FieldName should be marked as dirty")
		}
		if !product.Changes().Dirty(FieldDescription) {
			t.Error("FieldDescription should be marked as dirty")
		}
		if !product.Changes().Dirty(FieldCategory) {
			t.Error("FieldCategory should be marked as dirty")
		}
	})

	t.Run("fails to update archived product", func(t *testing.T) {
		now := time.Now()
		product2 := NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		product2.Archive(now)
		err := product2.UpdateDetails("New Name", "New Desc", "New Cat")
		if err != ErrProductAlreadyArchived {
			t.Errorf("UpdateDetails() error = %v, want ErrProductAlreadyArchived", err)
		}
	})

	t.Run("only marks changed fields as dirty", func(t *testing.T) {
		product3 := NewProduct("product-3", "Name", "Desc", "Cat", basePrice)
		product3.UpdateDetails("Name", "New Desc", "Cat")
		if product3.Changes().Dirty(FieldName) {
			t.Error("FieldName should not be marked as dirty when unchanged")
		}
		if !product3.Changes().Dirty(FieldDescription) {
			t.Error("FieldDescription should be marked as dirty")
		}
		if product3.Changes().Dirty(FieldCategory) {
			t.Error("FieldCategory should not be marked as dirty when unchanged")
		}
	})
}

func TestProduct_Activate(t *testing.T) {
	basePrice := NewMoney(10000, "USD")
	product := NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
	product.status = ProductStatusInactive

	t.Run("successfully activates product", func(t *testing.T) {
		err := product.Activate()
		if err != nil {
			t.Errorf("Activate() error = %v, want nil", err)
		}
		if product.Status() != ProductStatusActive {
			t.Errorf("Status() = %s, want %s", product.Status(), ProductStatusActive)
		}
		if !product.Changes().Dirty(FieldStatus) {
			t.Error("FieldStatus should be marked as dirty")
		}
		if len(product.DomainEvents()) != 1 {
			t.Errorf("DomainEvents() length = %d, want 1", len(product.DomainEvents()))
		}
		event := product.DomainEvents()[0]
		if event.EventName() != "product_activated" {
			t.Errorf("EventName() = %s, want product_activated", event.EventName())
		}
	})

	t.Run("no-op when already active", func(t *testing.T) {
		product2 := NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		product2.status = ProductStatusActive
		initialEvents := len(product2.DomainEvents())
		err := product2.Activate()
		if err != nil {
			t.Errorf("Activate() error = %v, want nil", err)
		}
		if len(product2.DomainEvents()) != initialEvents {
			t.Error("Should not emit event when already active")
		}
	})

	t.Run("fails to activate archived product", func(t *testing.T) {
		product3 := NewProduct("product-3", "Test", "Desc", "Cat", basePrice)
		now := time.Now()
		product3.Archive(now)
		err := product3.Activate()
		if err != ErrProductAlreadyArchived {
			t.Errorf("Activate() error = %v, want ErrProductAlreadyArchived", err)
		}
	})
}

func TestProduct_Deactivate(t *testing.T) {
	basePrice := NewMoney(10000, "USD")
	product := NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
	product.status = ProductStatusActive

	t.Run("successfully deactivates product", func(t *testing.T) {
		err := product.Deactivate()
		if err != nil {
			t.Errorf("Deactivate() error = %v, want nil", err)
		}
		if product.Status() != ProductStatusInactive {
			t.Errorf("Status() = %s, want %s", product.Status(), ProductStatusInactive)
		}
		if !product.Changes().Dirty(FieldStatus) {
			t.Error("FieldStatus should be marked as dirty")
		}
		if len(product.DomainEvents()) != 1 {
			t.Errorf("DomainEvents() length = %d, want 1", len(product.DomainEvents()))
		}
		event := product.DomainEvents()[0]
		if event.EventName() != "product_deactivated" {
			t.Errorf("EventName() = %s, want product_deactivated", event.EventName())
		}
	})

	t.Run("no-op when already inactive", func(t *testing.T) {
		product2 := NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		product2.status = ProductStatusInactive
		initialEvents := len(product2.DomainEvents())
		err := product2.Deactivate()
		if err != nil {
			t.Errorf("Deactivate() error = %v, want nil", err)
		}
		if len(product2.DomainEvents()) != initialEvents {
			t.Error("Should not emit event when already inactive")
		}
	})

	t.Run("fails to deactivate archived product", func(t *testing.T) {
		product3 := NewProduct("product-3", "Test", "Desc", "Cat", basePrice)
		now := time.Now()
		product3.Archive(now)
		err := product3.Deactivate()
		if err != ErrProductAlreadyArchived {
			t.Errorf("Deactivate() error = %v, want ErrProductAlreadyArchived", err)
		}
	})
}

func TestProduct_Archive(t *testing.T) {
	basePrice := NewMoney(10000, "USD")
	product := NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("successfully archives product", func(t *testing.T) {
		err := product.Archive(now)
		if err != nil {
			t.Errorf("Archive() error = %v, want nil", err)
		}
		// Note: archivedAt is not exposed via getter, but we can check events
		if !product.Changes().Dirty(FieldArchivedAt) {
			t.Error("FieldArchivedAt should be marked as dirty")
		}
		if len(product.DomainEvents()) != 1 {
			t.Errorf("DomainEvents() length = %d, want 1", len(product.DomainEvents()))
		}
		event := product.DomainEvents()[0]
		if event.EventName() != "product_archived" {
			t.Errorf("EventName() = %s, want product_archived", event.EventName())
		}
	})

	t.Run("fails to archive already archived product", func(t *testing.T) {
		product2 := NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		product2.Archive(now)
		err := product2.Archive(now.Add(1 * time.Hour))
		if err != ErrProductAlreadyArchived {
			t.Errorf("Archive() error = %v, want ErrProductAlreadyArchived", err)
		}
	})
}

func TestProduct_RemoveDiscount(t *testing.T) {
	basePrice := NewMoney(10000, "USD")
	product := NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	discount := &Discount{
		ID:        "discount-1",
		Amount:    NewMoney(10, 100),
		StartDate: startDate,
		EndDate:   endDate,
	}

	t.Run("successfully removes discount", func(t *testing.T) {
		product.status = ProductStatusActive
		product.ApplyDiscount(discount, now)
		if product.Discount() == nil {
			t.Fatal("Discount should be set before removal test")
		}

		err := product.RemoveDiscount()
		if err != nil {
			t.Errorf("RemoveDiscount() error = %v, want nil", err)
		}
		if product.Discount() != nil {
			t.Error("Discount() should be nil after removal")
		}
		if !product.Changes().Dirty(FieldDiscount) {
			t.Error("FieldDiscount should be marked as dirty")
		}
		if len(product.DomainEvents()) != 2 { // discount_applied + discount_removed
			t.Errorf("DomainEvents() length = %d, want 2", len(product.DomainEvents()))
		}
		lastEvent := product.DomainEvents()[len(product.DomainEvents())-1]
		if lastEvent.EventName() != "discount_removed" {
			t.Errorf("Last EventName() = %s, want discount_removed", lastEvent.EventName())
		}
	})

	t.Run("no-op when no discount exists", func(t *testing.T) {
		product2 := NewProduct("product-2", "Test", "Desc", "Cat", basePrice)
		initialEvents := len(product2.DomainEvents())
		err := product2.RemoveDiscount()
		if err != nil {
			t.Errorf("RemoveDiscount() error = %v, want nil", err)
		}
		if len(product2.DomainEvents()) != initialEvents {
			t.Error("Should not emit event when no discount exists")
		}
	})
}

func TestProduct_StateTransitions(t *testing.T) {
	basePrice := NewMoney(10000, "USD")
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("complete lifecycle transition", func(t *testing.T) {
		product := NewProduct("product-1", "Test", "Desc", "Cat", basePrice)
		
		// Start inactive
		if product.Status() != "" {
			t.Errorf("Initial Status() = %s, want empty", product.Status())
		}

		// Activate
		product.Activate()
		if product.Status() != ProductStatusActive {
			t.Errorf("Status() after Activate() = %s, want %s", product.Status(), ProductStatusActive)
		}

		// Deactivate
		product.Deactivate()
		if product.Status() != ProductStatusInactive {
			t.Errorf("Status() after Deactivate() = %s, want %s", product.Status(), ProductStatusInactive)
		}

		// Reactivate
		product.Activate()
		if product.Status() != ProductStatusActive {
			t.Errorf("Status() after second Activate() = %s, want %s", product.Status(), ProductStatusActive)
		}

		// Archive (final state)
		product.Archive(now)
		// After archiving, operations should fail
		err := product.Activate()
		if err != ErrProductAlreadyArchived {
			t.Errorf("Activate() after Archive() error = %v, want ErrProductAlreadyArchived", err)
		}
	})
}

func TestChangeTracker(t *testing.T) {
	t.Run("marks fields as dirty", func(t *testing.T) {
		ct := ChangeTracker{dirtyFields: make(map[string]bool)}
		ct.MarkDirty("field1")
		ct.MarkDirty("field2")

		if !ct.Dirty("field1") {
			t.Error("field1 should be dirty")
		}
		if !ct.Dirty("field2") {
			t.Error("field2 should be dirty")
		}
		if ct.Dirty("field3") {
			t.Error("field3 should not be dirty")
		}
	})

	t.Run("handles nil map", func(t *testing.T) {
		ct := ChangeTracker{}
		if ct.Dirty("field1") {
			t.Error("Dirty() should return false for nil map")
		}
		ct.MarkDirty("field1")
		if !ct.Dirty("field1") {
			t.Error("field1 should be dirty after MarkDirty")
		}
	})
}
