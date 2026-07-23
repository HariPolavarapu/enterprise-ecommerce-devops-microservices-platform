package domain

import "time"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusPacked    OrderStatus = "packed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              string
	CustomerID      string
	IdempotencyKey  string
	Status          OrderStatus
	TotalAmount     float64
	Currency        string
	OrderDate       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Items           []OrderItem
	PaymentHistory  []PaymentRecord
	TrackingNumber  string
	Version         int
}

type OrderItem struct {
	ID          string
	OrderID     string
	SKU         string
	Name        string
	Quantity    int
	UnitPrice   float64
	TotalPrice  float64
}

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusSuccess PaymentStatus = "success"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type PaymentRecord struct {
	ID           string
	OrderID      string
	Status       PaymentStatus
	Amount       float64
	Currency     string
	Provider     string
	TransactionID string
	CreatedAt    time.Time
}
