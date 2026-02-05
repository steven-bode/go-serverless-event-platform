package domain

import (
	"testing"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name        string
		orderID     OrderID
		customerID  CustomerID
		totalCents  int64
		expectError bool
		errorType   error
	}{
		{
			name:        "valid order",
			orderID:     "order-123",
			customerID:  "customer-456",
			totalCents:  10000,
			expectError: false,
		},
		{
			name:        "empty order id",
			orderID:     "",
			customerID:  "customer-456",
			totalCents:  10000,
			expectError: true,
			errorType:   ErrInvalidOrderID,
		},
		{
			name:        "empty customer id",
			orderID:     "order-123",
			customerID:  "",
			totalCents:  10000,
			expectError: true,
			errorType:   ErrInvalidCustomerID,
		},
		{
			name:        "zero total",
			orderID:     "order-123",
			customerID:  "customer-456",
			totalCents:  0,
			expectError: true,
			errorType:   ErrInvalidTotal,
		},
		{
			name:        "negative total",
			orderID:     "order-123",
			customerID:  "customer-456",
			totalCents:  -100,
			expectError: true,
			errorType:   ErrInvalidTotal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(tt.orderID, tt.customerID, tt.totalCents)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				if order != nil {
					t.Errorf("expected nil order, got %v", order)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if order == nil {
					t.Errorf("expected order but got nil")
					return
				}
				if order.ID != tt.orderID {
					t.Errorf("expected order ID %s, got %s", tt.orderID, order.ID)
				}
				if order.CustomerID != tt.customerID {
					t.Errorf("expected customer ID %s, got %s", tt.customerID, order.CustomerID)
				}
				if order.TotalCents != tt.totalCents {
					t.Errorf("expected total cents %d, got %d", tt.totalCents, order.TotalCents)
				}
			}
		})
	}
}
