package domain

import "time"

type Invoice struct {
	ID           string
	OrderID      string
	InvoiceNumber string
	Amount       float64
	Currency     string
	Status       string
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
