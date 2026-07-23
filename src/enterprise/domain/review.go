package domain

import "time"

type Review struct {
	ID          string
	CustomerID  string
	ProductSKU  string
	Rating      int
	Comment     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
}
