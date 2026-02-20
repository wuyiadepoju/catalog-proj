package domain

import (
	"testing"
	"time"
)

func TestDiscount_IsValidAt(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	discount := &Discount{
		ID:        "discount-1",
		Amount:    NewMoney(10, 100), // 10%
		StartDate: startDate,
		EndDate:   endDate,
	}

	tests := []struct {
		name     string
		checkTime time.Time
		expected bool
	}{
		{
			name:      "valid during discount period",
			checkTime: now,
			expected:  true,
		},
		{
			name:      "valid at start date",
			checkTime: startDate.Add(1 * time.Second),
			expected:  true,
		},
		{
			name:      "valid just before end date",
			checkTime: endDate.Add(-1 * time.Second),
			expected:  true,
		},
		{
			name:      "invalid before start date",
			checkTime: startDate.Add(-1 * time.Hour),
			expected:  false,
		},
		{
			name:      "invalid after end date",
			checkTime: endDate.Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "invalid at exact start date (not after)",
			checkTime: startDate,
			expected:  false,
		},
		{
			name:      "invalid at exact end date (not before)",
			checkTime: endDate,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discount.IsValidAt(tt.checkTime)
			if result != tt.expected {
				t.Errorf("IsValidAt(%v) = %v, want %v", tt.checkTime, result, tt.expected)
			}
		})
	}
}

func TestDiscount_IsValidAt_EdgeCases(t *testing.T) {
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	discount := &Discount{
		ID:        "discount-1",
		Amount:    NewMoney(10, 100),
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Test with same start and end date (should be invalid as now must be after start AND before end)
	sameDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	discountSameDate := &Discount{
		ID:        "discount-2",
		Amount:    NewMoney(10, 100),
		StartDate: sameDate,
		EndDate:   sameDate,
	}

	if discountSameDate.IsValidAt(sameDate) {
		t.Error("IsValidAt should return false when start and end dates are the same")
	}
}
