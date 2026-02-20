package domain

import (
	"math/big"
	"testing"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		expected string // as fraction string
	}{
		{
			name:     "creates money from cents",
			amount:   10000,
			currency: "USD",
			expected: "10000/100", // $100.00
		},
		{
			name:     "creates money with zero",
			amount:   0,
			currency: "USD",
			expected: "0/100",
		},
		{
			name:     "creates money with decimal cents",
			amount:   12345,
			currency: "USD",
			expected: "12345/100", // $123.45
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money := NewMoney(tt.amount, tt.currency)
			if money.String() != tt.expected {
				t.Errorf("NewMoney(%d, %s) = %s, want %s", tt.amount, tt.currency, money.String(), tt.expected)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	m1 := NewMoney(10000, "USD") // $100.00
	m2 := NewMoney(5000, "USD")  // $50.00

	result := Add(m1, m2)
	expected := big.NewRat(15000, 100) // $150.00

	if result.Cmp(expected) != 0 {
		t.Errorf("Add($100.00, $50.00) = %s, want %s", result.String(), expected.String())
	}
}

func TestSubtract(t *testing.T) {
	m1 := NewMoney(10000, "USD") // $100.00
	m2 := NewMoney(3000, "USD")  // $30.00

	result := Subtract(m1, m2)
	expected := big.NewRat(7000, 100) // $70.00

	if result.Cmp(expected) != 0 {
		t.Errorf("Subtract($100.00, $30.00) = %s, want %s", result.String(), expected.String())
	}
}

func TestMultiply(t *testing.T) {
	m1 := NewMoney(10000, "USD") // $100.00
	m2 := NewMoney(20, 100)      // 0.20 (20%)

	result := Multiply(m1, m2)
	expected := big.NewRat(2000, 100) // $20.00

	if result.Cmp(expected) != 0 {
		t.Errorf("Multiply($100.00, 0.20) = %s, want %s", result.String(), expected.String())
	}
}

func TestMoneyOperations(t *testing.T) {
	// Test complex calculation: (100 + 50) * 0.10
	base := NewMoney(10000, "USD")
	addend := NewMoney(5000, "USD")
	percentage := NewMoney(10, 100) // 10%

	sum := Add(base, addend)
	result := Multiply(sum, percentage)
	expected := big.NewRat(1500, 100) // $15.00

	if result.Cmp(expected) != 0 {
		t.Errorf("Complex calculation = %s, want %s", result.String(), expected.String())
	}
}
