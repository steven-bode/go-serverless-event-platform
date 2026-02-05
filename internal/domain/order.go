package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidOrderID     = errors.New("invalid order id")
	ErrInvalidCustomerID  = errors.New("invalid customer id")
	ErrInvalidTotal       = errors.New("invalid total: must be greater than 0")
	ErrOrderAlreadyExists = errors.New("order already exists")
)

type OrderID string
type CustomerID string

type Order struct {
	ID         OrderID
	CustomerID CustomerID
	TotalCents int64
	CreatedAt  time.Time
}

func NewOrder(id OrderID, customerID CustomerID, totalCents int64) (*Order, error) {
	if err := ValidateOrderID(id); err != nil {
		return nil, err
	}
	if err := ValidateCustomerID(customerID); err != nil {
		return nil, err
	}
	if err := ValidateTotal(totalCents); err != nil {
		return nil, err
	}

	return &Order{
		ID:         id,
		CustomerID: customerID,
		TotalCents: totalCents,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func ValidateOrderID(id OrderID) error {
	if id == "" {
		return ErrInvalidOrderID
	}
	return nil
}

func ValidateCustomerID(customerID CustomerID) error {
	if customerID == "" {
		return ErrInvalidCustomerID
	}
	return nil
}

func ValidateTotal(totalCents int64) error {
	if totalCents <= 0 {
		return ErrInvalidTotal
	}
	return nil
}
